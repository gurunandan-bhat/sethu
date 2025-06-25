package payment

import "fmt"

type Card struct {
	EMI           bool
	Entity        string
	ID            string
	International bool
	Issuer        string
	Last4         string
	Name          string
	Network       string
	SubType       string `mapstructure:"sub_type"`
	TokenIIN      string `mapstructure:"token_iin"`
	Type          string
}

type UPI struct {
	VPA string
}

type AcquirerData struct {
	AuthCode         string `mapstructure:"auth_code"`
	RRN              string
	UPITransactionID string `mapstructure:"upi_transaction_id"`
}

type Notes struct {
	Name    string
	Project string
	Email   string
}

type PaymentError struct {
	Code        string      `mapstructure:"error_code" schema:"error[code]"`
	Description string      `mapstructure:"error_description" schema:"error[description]"`
	Reason      string      `mapstructure:"error_reason" schema:"error[reason]"`
	Source      string      `mapstructure:"error_source" schema:"error[source]"`
	Step        string      `mapstructure:"error_step" schema:"error[step]"`
	Metadata    PaymentInfo `mapstructure:"error_metadata" schema:"error[metadata]"`
}

func (pe PaymentError) Error() string {

	return fmt.Sprintf("Payment failed with code %s: %s (%s:%s)", pe.Code, pe.Description, pe.Source, pe.Reason)
}

type PaymentInfo struct {
	PaymentID string `schema:"razorpay_payment_id" mapstructure:"payment_id" json:"payment_id,omitempty"`
	OrderID   string `schema:"razorpay_order_id" mapstructure:"order_id" json:"order_id,omitempty"`
	Signature string `schema:"razorpay_signature"`
}

type Payment struct {
	AcquirerData  AcquirerData `mapstructure:"acquirer_data"`
	Amount        float64
	AmountINR     string
	Bank          string
	Captured      bool
	Card          Card
	CardID        string `mapstructure:"card_id"`
	Contact       string
	CreatedAt     int64 `mapstructure:"created_at"`
	Currency      string
	Description   string
	Email         string
	Entity        string
	Error         PaymentError `mapstructure:",squash"`
	Fee           float64
	ID            string
	International bool
	InvoiceID     string `mapstructure:"invoice_id"`
	Method        string
	Notes         Notes
	OrderID       string `mapstructure:"order_id"`
	RefundStatus  string `mapstructure:"refund_status"`
	Status        string
	Tax           float64
	UPI           UPI
	VPA           string
	Wallet        string
}
