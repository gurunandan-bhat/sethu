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
)

type Donate struct {
	Name      string  `schema:"name,required"`
	EMail     string  `schema:"email,required"`
	AmountINR float64 `schema:"amount,required"`
	Project   string  `schema:"project,required"`
}

var decoder = schema.NewDecoder()

func (s *Service) order(w http.ResponseWriter, r *http.Request) error {

	fmt.Println("Entered Order Handler")

	cfg, err := config.Configuration()
	if err != nil {
		fmt.Println("Config: ", err)
		return fmt.Errorf("unable to read configuration: %w", err)
	}

	if err := r.ParseMultipartForm(10 * 1024); err != nil {
		fmt.Println("Parse Form:", err)
		return fmt.Errorf("error parsing form: %w", err)
	}

	var donate Donate
	if err := decoder.Decode(&donate, r.Form); err != nil {
		fmt.Println("Decoder error: ", err)
		return fmt.Errorf("error decoding form data: %w", err)
	}

	keyID := cfg.RazorPay.KeyID
	keySecret := cfg.RazorPay.KeySecret
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
		fmt.Println(err)
		return fmt.Errorf("error creating order: %w", err)
	}
	fmt.Printf("%+v\n", body)

	vRzpOrderID, ok := body["id"].(string)
	if !ok {
		fmt.Println("order_id is not a string")
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
		fmt.Println(err)
		return err
	}
	order.VRzpKeyID = keyID

	jsonBytes, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	fmt.Printf("%+v\n", string(jsonBytes))
	return s.renderJSON(w, jsonBytes, http.StatusOK)
}
