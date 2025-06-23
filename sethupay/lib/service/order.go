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
	PAN       string  `schema:"pan"`
}

var decoder = schema.NewDecoder()

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
			"project": donate.Project,
			"name":    donate.Name,
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

	payResponse, payError := getPaymentResponse(r)
	if payError != nil {
		return payError
	}

	params := map[string]any{
		"razorpay_order_id":   payResponse.OrderID,
		"razorpay_payment_id": payResponse.PaymentID,
	}

	secret, err := s.RazorpaySecret()
	if err != nil {
		return err
	}

	matched := utils.VerifyPaymentSignature(params, payResponse.Signature, secret.KeySecret)
	if !matched {
		return errors.New("signature mismatch, aborting payment")
	}

	// Payment was successful. Handle success
	client := razorpay.NewClient(secret.KeyID, secret.KeySecret)
	details, err := client.Payment.Fetch(payResponse.PaymentID, nil, nil)
	if details["error_code"] != nil {
		return errors.New("error fetching payment details " + err.Error())
	}

	var paymentData payment.Payment
	if err := mapstructure.Decode(details, &paymentData); err != nil {
		return fmt.Errorf("error encoding response: %w", err)
	}

	s.sendEmail(paymentData.Email, "payment-email.go.html", paymentData)
	return s.render(w, "thank-you.go.html", paymentData, http.StatusOK)
}

func getPaymentResponse(r *http.Request) (payment.PaymentInfo, error) {

	defer r.Body.Close()
	rsp := payment.PaymentInfo{}

	if err := r.ParseForm(); err != nil {
		return rsp, fmt.Errorf("error parsing form: %w", err)
	}
	data := r.PostForm
	// Check if the payment has failed
	if data.Has("error[code]") {
		// We have an error so we should parse it and return
		meta := payment.PaymentInfo{}
		if data.Has("error[metadata]") {
			jsonBytes := []byte(data.Get("error[metadata]"))
			if err := json.Unmarshal(jsonBytes, &meta); err != nil {
				return rsp, fmt.Errorf("error unmarshaling payment response: %w", err)
			}
		}
		err := payment.PaymentError{
			Code:        data.Get("error[code]"),
			Description: data.Get("error[description]"),
			Reason:      data.Get("error[reason]"),
			Source:      data.Get("error[source]"),
			Step:        data.Get("err"),
			Metadata:    meta,
		}
		return rsp, err
	}

	// No error so just decode the payment info
	if err := decoder.Decode(&rsp, data); err != nil {
		return rsp, fmt.Errorf("error decoding form values: %w", err)
	}

	return rsp, nil
}
