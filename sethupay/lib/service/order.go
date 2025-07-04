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

type Donate struct {
	Name      string  `schema:"name,required"`
	EMail     string  `schema:"email,required"`
	AmountINR float64 `schema:"amount,required"`
	Project   string  `schema:"project,required"`
	Address1  string  `schema:"addr1,required"`
	Address2  string  `schema:"addr2"`
	City      string  `schema:"city,required"`
	Pin       string  `schema:"pin,required"`
	State     string  `schema:"state,required"`
	PAN       string  `schema:"pan"`
}

var decoder = schema.NewDecoder()

// This is called from JS so we output a JSON and return an error
func (s *Service) order(w http.ResponseWriter, r *http.Request) error {

	if err := r.ParseMultipartForm(10 * 1024); err != nil {
		return fmt.Errorf("error parsing form: %w", err)
	}
	defer r.Body.Close()

	var donate Donate
	if err := decoder.Decode(&donate, r.Form); err != nil {
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

	data := map[string]any{
		"amount":          amountINR,
		"currency":        "INR",
		"receipt":         reciept,
		"partial_payment": false,
		"notes": map[string]any{
			"Project":  donate.Project,
			"Name":     donate.Name,
			"Address1": donate.Address1,
			"Address2": donate.Address2,
			"City":     donate.City,
			"Pin":      donate.Pin,
			"State":    donate.State,
			"PAN":      donate.PAN,
		},
	}

	body, err := client.Order.Create(data, nil)
	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}

	vRzpOrderID, ok := body["id"].(string)
	if !ok {
		return fmt.Errorf("cannot read order_id as string: %w", err)
	}

	order := model.DBOrder{
		VRcptID:     reciept,
		VRzpOrderID: vRzpOrderID,
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
		return err
	}

	order.VRzpKeyID = keyID
	jsonBytes, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	return s.renderJSON(w, jsonBytes, http.StatusOK)
}

func (s *Service) paid(w http.ResponseWriter, r *http.Request) error {

	paymentResponse, payError := s.getPaymentResponse(r)
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
		return fmt.Errorf("error encoding response: %w", err)
	}

	if err := s.Model.LogPaymentStatus(paymentResponse, "Paid", fmt.Sprintf("%+v\n", details)); err != nil {
		return fmt.Errorf("error updating order after payment: %w", err)
	}

	paymentData.AmountINR = fmt.Sprintf("%.2f", (paymentData.Amount / 100.00))

	emailTmpl := "default-success.go.html"
	s.sendEmail(paymentData.Email, emailTmpl, paymentData)

	return s.render(w, "thank-you.go.html", paymentData, http.StatusOK)
}

func (s *Service) getPaymentResponse(r *http.Request) (payment.PaymentResponse, error) {

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
