package pay

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRecentOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var request map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		assert.Equal(t, "receipt", request["receipt-data"])
		_, _ = w.Write([]byte(`{"status":0,"receipt":{"bundle_id":"com.example.app","in_app":[{"transaction_id":"old","purchase_date_ms":"1"},{"transaction_id":"new","purchase_date_ms":"2"}]}}`))
	}))
	defer server.Close()

	pay := New(Config{
		Endpoint: server.URL,
		BundleID: "com.example.app",
	})
	iap, err := pay.GetRecentOrder("receipt")
	require.NoError(t, err)
	assert.Equal(t, "new", iap.TransactionID)
}
