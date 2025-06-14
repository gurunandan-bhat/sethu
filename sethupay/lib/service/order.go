package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sethupay/lib/config"
	"sethupay/lib/model"

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
}

type Payment struct {
	PaymentID string `schema:"razorpay_payment_id"`
	OrderID   string `schema:"razorpay_order_id"`
	Signature string `schema:"razorpay_signature"`
}

type PaymentError struct {
	Code        string `json:"error_code,omitempty" schema:"error_code"`
	Description string `json:"error_description,omitempty" schema:"error_description"`
	Reason      string `json:"error_reason,omitempty" schema:"error_reason"`
	Source      string `json:"error_source,omitempty" schema:"error_source"`
	Step        string `json:"error_step,omitempty" schema:"error_step"`
}

type PaymentResponse struct {
	Project  string
	Name     string
	Currency string
	Amount   string
}

func (pe PaymentError) Error() string {

	return fmt.Sprintf("error_code %s; %s %s %s", pe.Code, pe.Description, pe.Reason, pe.Source)
}

var decoder = schema.NewDecoder()

func (s *Service) order(w http.ResponseWriter, r *http.Request) error {

	cfg, err := config.Configuration()
	if err != nil {
		return fmt.Errorf("unable to read configuration: %w", err)
	}

	if err := r.ParseMultipartForm(10 * 1024); err != nil {
		return fmt.Errorf("error parsing form: %w", err)
	}
	defer r.Body.Close()

	var donate Donate
	if err := decoder.Decode(&donate, r.Form); err != nil {
		return fmt.Errorf("error decoding form data: %w", err)
	}

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
			"payer":   donate.Name,
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

	status := Payment{}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("error parsing form: %w", err)
	}

	// Check if the payment has failed
	if err := r.PostFormValue("error[code]"); err != "" {
		return errors.New(r.PostFormValue("error[reason]"))
	}

	if err := decoder.Decode(&status, r.PostForm); err != nil {
		return fmt.Errorf("error decoding form values: %w", err)
	}

	// Check signature
	cfg, err := config.Configuration()
	if err != nil {
		return fmt.Errorf("error fetching configuration: %w", err)
	}

	// Generate the expected signature
	key := cfg.RazorPay.Test
	if cfg.InProduction {
		key = cfg.RazorPay.Live
	}
	params := map[string]any{
		"razorpay_order_id":   status.OrderID,
		"razorpay_payment_id": status.PaymentID,
	}
	matched := utils.VerifyPaymentSignature(params, status.Signature, key.KeySecret)
	if !matched {
		return errors.New("signature mismatch, aborting payment")
	}

	// Payment was successful. Handle success
	client := razorpay.NewClient(key.KeyID, key.KeySecret)
	details, err := client.Payment.Fetch(status.PaymentID, nil, nil)
	if details["error_code"] != nil {
		return errors.New("error fetching payment details " + err.Error())
	}
	tmplData, err := paymentData(details)
	if err != nil {
		return fmt.Errorf("error encoding response: %w", err)
	}
	// jsonBytes, err := json.Marshal(details)
	// if err != nil {
	// 	return fmt.Errorf("error marshaling order from session: %w", err)
	// }
	// fmt.Println(details)

	// return s.renderJSON(w, jsonBytes, http.StatusOK)

	return s.render(w, "thank-you.go.html", tmplData, http.StatusOK)
}

func paymentData(details map[string]any) (PaymentResponse, error) {

	pData := PaymentResponse{}

	notes, ok := details["notes"].(map[string]any)
	if !ok {
		return pData, errors.New("key notes of details has unexpected type")
	}
	project, ok := notes["project"].(string)
	if !ok {
		return pData, errors.New("key project of notes has unexpected type")
	}
	pData.Project = project

	payer, ok := notes["payer"].(string)
	if !ok {
		return pData, errors.New("key project of notes has unexpected type")
	}
	pData.Name = payer

	currency, ok := details["currency"].(string)
	if !ok {
		return pData, errors.New("key currency of details has unexpected type")
	}
	pData.Currency = currency

	amt, ok := details["amount"].(float64)
	if !ok {
		return pData, errors.New("key amount of details has unexpected type")
	}
	amount := fmt.Sprintf("%.2f", (amt / 100))
	pData.Amount = amount

	return pData, nil
}

// func mkPaymentError(details map[string]any) PaymentError {

// 	m := make(map[string]any)
// 	for key, val := range details {
// 		if strings.HasPrefix(key, "error_") {
// 			m[key] = val
// 		}
// 	}

// 	return PaymentError{}

// }

// func mkTypeError(field string) PaymentError {

// 	return PaymentError{
// 		"curom error",
// 		fmt.Sprintf("incorrect type %s in payment details", field),
// 		"", "", "",
// 	}
// }
