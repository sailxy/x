package pay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRecentOrder(t *testing.T) {
	pay := New(Config{
		Endpoint: "https://sandbox.itunes.apple.com/verifyReceipt",
		BundleID: "",
	})
	iap, err := pay.GetRecentOrder("")
	assert.NoError(t, err)
	t.Log(iap)
}
