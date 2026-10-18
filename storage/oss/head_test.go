package oss

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	aliyunoss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/stretchr/testify/require"
)

func TestHeadObject_ConfiguredBucketMetadataAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		size   string
	}{{"exists", 200, "123"}, {"missing", 404, "0"}, {"unavailable", 503, "0"}, {"invalid size", 200, ""}} {
		t.Run(tc.name, func(t *testing.T) {
			client := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodHead, r.Method)
				require.Equal(t, "/private-test/proof.pdf", r.URL.Path)
				if tc.size != "" {
					w.Header().Set("Content-Length", tc.size)
				}
				w.Header().Set("Content-Type", "application/pdf")
				w.WriteHeader(tc.status)
			})
			meta, err := client.HeadObject(context.Background(), "proof.pdf")
			if tc.name == "exists" {
				require.NoError(t, err)
				require.Equal(t, &ObjectMeta{Size: 123, ContentType: "application/pdf"}, meta)
			} else {
				require.Error(t, err)
				require.Nil(t, meta)
			}
			if tc.status == 404 {
				var original *aliyunoss.ServiceError
				require.True(t, errors.As(err, &original))
				require.Equal(t, 404, original.StatusCode)
			}
		})
	}
}

func TestHeadObject_ContextAndEmptyKey(t *testing.T) {
	client := localOSSClient(t, func(w http.ResponseWriter, _ *http.Request) { _, err := fmt.Fprint(w, ""); require.NoError(t, err) })
	_, err := client.HeadObject(context.Background(), "")
	require.Error(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.HeadObject(ctx, "proof.pdf")
	require.ErrorIs(t, err, context.Canceled)
}
