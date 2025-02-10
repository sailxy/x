package pay

import (
	"errors"
	"fmt"
)

const StatusOK = 0

type InApp struct {
	Quantity                string `json:"quantity"`
	ProductID               string `json:"product_id"`
	TransactionID           string `json:"transaction_id"`
	OriginalTransactionID   string `json:"original_transaction_id"`
	PurchaseDate            string `json:"purchase_date"`
	PurchaseDateMs          string `json:"purchase_date_ms"`
	PurchaseDatePst         string `json:"purchase_date_pst"`
	OriginalPurchaseDate    string `json:"original_purchase_date"`
	OriginalPurchaseDateMs  string `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePst string `json:"original_purchase_date_pst"`
	IsTrialPeriod           string `json:"is_trial_period"`
	InAppOwnershipType      string `json:"in_app_ownership_type"`
}

type ReceiptPayload struct {
	Receipt struct {
		ReceiptType                string  `json:"receipt_type"`
		AdamID                     int     `json:"adam_id"`
		AppItemID                  int     `json:"app_item_id"`
		BundleID                   string  `json:"bundle_id"`
		ApplicationVersion         string  `json:"application_version"`
		DownloadID                 int     `json:"download_id"`
		VersionExternalIdentifier  int     `json:"version_external_identifier"`
		ReceiptCreationDate        string  `json:"receipt_creation_date"`
		ReceiptCreationDateMs      string  `json:"receipt_creation_date_ms"`
		ReceiptCreationDatePst     string  `json:"receipt_creation_date_pst"`
		RequestDate                string  `json:"request_date"`
		RequestDateMs              string  `json:"request_date_ms"`
		RequestDatePst             string  `json:"request_date_pst"`
		OriginalPurchaseDate       string  `json:"original_purchase_date"`
		OriginalPurchaseDateMs     string  `json:"original_purchase_date_ms"`
		OriginalPurchaseDatePst    string  `json:"original_purchase_date_pst"`
		OriginalApplicationVersion string  `json:"original_application_version"`
		InApp                      []InApp `json:"in_app"`
	} `json:"receipt"`
	Environment string `json:"environment"`
	Status      int    `json:"status"`
}

func (r *ReceiptPayload) check(bundleID string) error {
	if r.Status != StatusOK {
		return fmt.Errorf("%d", r.Status)
	}
	if r.Receipt.BundleID != bundleID {
		return errors.New("invalid receipt")
	}
	if len(r.Receipt.InApp) == 0 {
		return errors.New("order not found")
	}

	return nil
}

// Get the most recent order.
func (r *ReceiptPayload) RecentOrder(bundleID string) (*InApp, error) {
	if err := r.check(bundleID); err != nil {
		return nil, err
	}

	inApp := r.Receipt.InApp[0]
	for _, v := range r.Receipt.InApp {
		if v.PurchaseDateMs > inApp.PurchaseDateMs {
			inApp = v
		}
	}
	return &inApp, nil
}
