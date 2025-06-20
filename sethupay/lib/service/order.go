package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

	defer r.Body.Close()

	response := payment.PaymentResponse{}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("error parsing form: %w", err)
	}

	// Check if the payment has failed
	if err := r.PostFormValue("error[code]"); err != "" {
		return errors.New(s.logPaymentError(r.PostForm))
	}

	if err := decoder.Decode(&response, r.PostForm); err != nil {
		return fmt.Errorf("error decoding form values: %w", err)
	}

	// Generate the expected signature
	cfg := s.Config
	key := cfg.RazorPay.Test
	if cfg.InProduction {
		key = cfg.RazorPay.Live
	}
	params := map[string]any{
		"razorpay_order_id":   response.OrderID,
		"razorpay_payment_id": response.PaymentID,
	}
	matched := utils.VerifyPaymentSignature(params, response.Signature, key.KeySecret)
	if !matched {
		return errors.New("signature mismatch, aborting payment")
	}

	// Payment was successful. Handle success
	client := razorpay.NewClient(key.KeyID, key.KeySecret)
	details, err := client.Payment.Fetch(response.PaymentID, nil, nil)
	if details["error_code"] != nil {
		return errors.New("error fetching payment details " + err.Error())
	}

	var paymentData payment.Payment
	if err := mapstructure.Decode(details, &paymentData); err != nil {
		return fmt.Errorf("error encoding response: %w", err)
	}

	return s.render(w, "thank-you.go.html", paymentData, http.StatusOK)
}

func (s *Service) logPaymentError(err url.Values) string {

	meta := payment.PaymentResponse{}
	if err.Has("error[metadata]") {
		if err := json.Unmarshal([]byte(err.Get("error[metadata]")), &meta); err != nil {
			return fmt.Sprint("error unmarshaling payment response: ", err.Error())
		}
	}
	code := err.Get("error[code]")
	description := err.Get("error[description]")
	reason := err.Get("error[reason]")

	errStr := fmt.Sprintf("%s: %s (%s)", code, description, reason)
	if err := s.Model.LogPaymentStatus(meta, "Failed", errStr); err != nil {
		return "error logging response for " + errStr
	}

	return "Payment Failed"
}
