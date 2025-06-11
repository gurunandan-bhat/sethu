package service

import (
	"encoding/json"
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

	s.setSessionVar(r, w, "orderID", vRzpOrderID)

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
	// orderID, err := s.getSessionVar(r, "orderID")
	// if err != nil {
	// 	return fmt.Errorf("error fetching orderID: %w", err)
	// }

	// client := razorpay.NewClient(key.KeyID, key.KeySecret)
	params := map[string]any{
		"razorpay_order_id":   status.OrderID,
		"razorpay_payment_id": status.PaymentID,
	}

	matched := utils.VerifyPaymentSignature(params, status.Signature, key.KeySecret)
	fmt.Println(matched)

	return nil
}
