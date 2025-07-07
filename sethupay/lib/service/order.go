package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sethupay/lib/model"
	"sethupay/lib/payment"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	razorpay "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/utils"
)

var decoder = schema.NewDecoder()

// This is called from JS so we output a JSON and return an error
func (s *Service) order(w http.ResponseWriter, r *http.Request) error {

	if err := r.ParseMultipartForm(10 * 1024); err != nil {
		s.Logger.Error("Validating form: ", "error", err.Error())
		errJSON := fmt.Sprintf(`{"error": "%s"}`, err.Error())
		s.renderJSON(w, []byte(errJSON), http.StatusBadRequest)
		return nil
	}
	defer r.Body.Close()

	var donate payment.Notes
	if err := decoder.Decode(&donate, r.Form); err != nil {
		fmt.Printf("Found type to be: %T\n", err)
		missingErr, ok := err.(schema.MultiError)
		if ok {
			jsonBytes, err := json.Marshal(missingErr)
			if err != nil {
				s.Logger.Error("Unmarshaling schema error: ", "error", err.Error())
				errJSON := fmt.Sprintf(`{"error": "%s"}`, err.Error())
				s.renderJSON(w, []byte(errJSON), http.StatusBadRequest)
				return nil
			}
			fmt.Printf("%+v\n", missingErr)
			s.renderJSON(w, jsonBytes, http.StatusBadRequest)
			return nil
		}
		s.Logger.Error("Validating form: ", "error", err.Error())
		return fmt.Errorf("error decoding form data: %w", err)
	}

	cfg := s.Config
	key := cfg.RazorPay.Test
	if cfg.InProduction {
		key = cfg.RazorPay.Live
	}
	keyID := key.KeyID
	keySecret := key.KeySecret

	client := razorpay.NewClient(keyID, keySecret)
	id := uuid.New()
	reciept := id.String()
	amountINR := int(100 * donate.AmountINR)

	var notes = new(map[string]any)
	if err := mapstructure.Decode(donate, notes); err != nil {
		s.Logger.Error("error converting struct to map: ", "error", err.Error())
		errJSON := fmt.Sprintf(`{"error": "%s"}`, err.Error())
		s.renderJSON(w, []byte(errJSON), http.StatusBadRequest)
		return nil
	}
	data := map[string]any{
		"amount":          amountINR,
		"currency":        "INR",
		"receipt":         reciept,
		"partial_payment": false,
		"notes":           *notes,
	}
	s.Logger.Info("Notes encoded to map: ", "data", data)

	body, err := client.Order.Create(data, nil)
	if err != nil {
		s.Logger.Error("error creating order: ", "error", err.Error())
		errJSON := fmt.Sprintf(`{"error": "%s"}`, err.Error())
		s.renderJSON(w, []byte(errJSON), http.StatusBadRequest)
		return nil
	}

	vRzpOrderID, ok := body["id"].(string)
	if !ok {
		s.Logger.Error("error - no razorpay id found: ", "error", "no field id in response body")
		errJSON := fmt.Sprintf(`{"error": "%s"}`, "no field id in response body")
		s.renderJSON(w, []byte(errJSON), http.StatusBadRequest)
		return nil
	}

	order := model.DBOrder{
		VRzpOrderID: vRzpOrderID,
		VRcptID:     reciept,
		VName:       donate.Name,
		VEmail:      donate.EMail,
		IAmount:     amountINR,
		VProject:    donate.Project,
		VAddress1:   donate.Address1,
		VAddress2:   donate.Address2,
		VCity:       donate.City,
		VPin:        donate.Pin,
		VState:      donate.State,
		VPAN:        donate.PAN,
		VStatus:     "Created",
	}
	if err := s.Model.NewOrder(&order); err != nil {
		s.Logger.Error("error creating db order: ", "error", err.Error())
		errJSON := fmt.Sprintf(`{"error": "%s"}`, err.Error())
		s.renderJSON(w, []byte(errJSON), http.StatusBadRequest)
		return nil
	}

	order.VRzpKeyID = keyID
	jsonBytes, err := json.Marshal(order)
	if err != nil {
		s.Logger.Error("error marshaling order from response: ", "error", err.Error())
		errJSON := fmt.Sprintf(`{"error": "%s"}`, err.Error())
		s.renderJSON(w, []byte(errJSON), http.StatusBadRequest)
		return nil
	}

	return s.renderJSON(w, jsonBytes, http.StatusOK)
}

func (s *Service) paid(w http.ResponseWriter, r *http.Request) error {

	paymentResponse, payError := s.decodePaymentResponse(r)
	if payError != nil {
		return payError
	}

	params := map[string]any{
		"razorpay_order_id":   paymentResponse.OrderID,
		"razorpay_payment_id": paymentResponse.PaymentID,
	}

	secret, err := s.RazorpaySecret()
	if err != nil {
		return err
	}

	matched := utils.VerifyPaymentSignature(params, paymentResponse.Signature, secret.KeySecret)
	if !matched {
		return errors.New("signature mismatch, aborting payment")
	}

	// Payment was successful. Handle success
	client := razorpay.NewClient(secret.KeyID, secret.KeySecret)
	details, err := client.Payment.Fetch(paymentResponse.PaymentID, nil, nil)
	if details["error_code"] != nil {
		return errors.New("error fetching payment details " + err.Error())
	}

	var paymentData payment.Payment
	if err := mapstructure.Decode(details, &paymentData); err != nil {
		// There was an error which could be due to decoding or a
		// payment error, which can checked with error_code. We
		// therefore need to do two things - Log the error and send
		// error apology page and the error email.
		return fmt.Errorf("error encoding response: %w", err)
	}

	if err := s.Model.LogPaymentStatus(paymentResponse, "Paid", fmt.Sprintf("%+v\n", details)); err != nil {
		return fmt.Errorf("error updating order after payment: %w", err)
	}

	// Convert paise to rupees and send success email
	paymentData.AmountINR = fmt.Sprintf("%.2f", (paymentData.Amount / 100.00))
	//emailTmpl := "default-success.go.html"
	// s.sendEmail(paymentData.Email, emailTmpl, paymentData)
	return s.render(w, "thank-you.go.html", paymentData, http.StatusOK)
}

func (s *Service) decodePaymentResponse(r *http.Request) (payment.PaymentResponse, error) {

	defer r.Body.Close()
	rsp := payment.PaymentResponse{}

	if err := r.ParseForm(); err != nil {
		return rsp, fmt.Errorf("error parsing form: %w", err)
	}
	data := r.PostForm
	// Check if the payment has failed
	if data.Has("error[code]") {
		// We have an error so we should parse it and return
		if data.Has("error[metadata]") {
			jsonBytes := []byte(data.Get("error[metadata]"))
			if err := json.Unmarshal(jsonBytes, &rsp); err != nil {
				return rsp, fmt.Errorf("error unmarshaling payment response: %w", err)
			}
		}
		err := payment.PaymentError{
			Code:        data.Get("error[code]"),
			Description: data.Get("error[description]"),
			Reason:      data.Get("error[reason]"),
			Source:      data.Get("error[source]"),
			Step:        data.Get("err"),
			Metadata:    rsp,
		}
		if err := s.Model.LogPaymentStatus(rsp, "Failed", fmt.Sprintf("%+v\n", data)); err != nil {
			return rsp, fmt.Errorf("error updating order after payment: %w", err)
		}

		return rsp, err
	}

	// No error so just decode the payment info
	if err := decoder.Decode(&rsp, data); err != nil {
		return rsp, fmt.Errorf("error decoding form values: %w", err)
	}

	return rsp, nil
}
