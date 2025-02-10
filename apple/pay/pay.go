package pay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Config struct {
	Endpoint string
	BundleID string
}

type Pay struct {
	endpoint string
	bundleID string
}

func New(c Config) *Pay {
	return &Pay{
		endpoint: c.Endpoint,
		bundleID: c.BundleID,
	}
}

// IAP
func (p *Pay) GetRecentOrder(receipt string) (*InApp, error) {
	data := make(map[string]interface{})
	data["receipt-data"] = receipt
	bytesData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(p.endpoint, "application/json", bytes.NewReader(bytesData))
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload ReceiptPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		return nil, err
	}

	return payload.RecentOrder(p.bundleID)
}
