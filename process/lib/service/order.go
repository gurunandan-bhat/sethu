package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"process/lib/config"

	razorpay "github.com/razorpay/razorpay-go"
)

func (s *Service) NewOrder(w http.ResponseWriter, r *http.Request) error {

	cfg, err := config.Configuration()
	if err != nil {
		return fmt.Errorf("unable to read configuration: %w", err)
	}

	keyID := cfg.RazorPay.KeyID
	keySecret := cfg.RazorPay.KeySecret
	client := razorpay.NewClient(keyID, keySecret)

	data := map[string]any{
		"amount":          500,
		"currency":        "INR",
		"receipt":         "testRcptID",
		"partial_payment": false,
		"notes": map[string]any{
			"testKey1": "testValue1",
			"testKey2": "testValue2",
		},
	}

	body, err := client.Order.Create(data, nil)
	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	return s.renderJSON(w, jsonBytes, http.StatusOK)
}
