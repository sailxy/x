package oss

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	aliyunoss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/stretchr/testify/require"
)

func localOSSClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	sdk := aliyunoss.NewClient(aliyunoss.LoadDefaultConfig().WithEndpoint(server.URL).WithRegion("cn-hangzhou").WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test-ak", "test-secret")).WithUsePathStyle(true).WithRetryMaxAttempts(1))
	return &Client{sdkClient: sdk, bucketName: "private-test"}
}

func TestBucketState_ReadsWithoutMutationAndHonorsCancellation(t *testing.T) {
	queries := 0
	client := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
		queries++
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/private-test", strings.TrimSuffix(r.URL.Path, "/"))
		w.Header().Set("Content-Type", "application/xml")
		if r.URL.Query().Has("acl") {
			_, err := fmt.Fprint(w, `<AccessControlPolicy><AccessControlList><Grant>private</Grant></AccessControlList></AccessControlPolicy>`)
			require.NoError(t, err)
		} else {
			_, err := fmt.Fprint(w, `<VersioningConfiguration/>`)
			require.NoError(t, err)
		}
	})
	state, err := client.BucketState(context.Background())
	require.NoError(t, err)
	require.Equal(t, &BucketState{ACL: "private"}, state)
	require.Equal(t, 2, queries)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.BucketState(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestBucketState_ReturnsPublicAndVersionStatesWithoutBusinessRules(t *testing.T) {
	for _, status := range []string{"Enabled", "Suspended", "unknown"} {
		t.Run(status, func(t *testing.T) {
			client := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
				body := `<VersioningConfiguration><Status>` + status + `</Status></VersioningConfiguration>`
				if r.URL.Query().Has("acl") {
					body = `<AccessControlPolicy><AccessControlList><Grant>public-read</Grant></AccessControlList></AccessControlPolicy>`
				}
				_, err := fmt.Fprint(w, body)
				require.NoError(t, err)
			})
			state, err := client.BucketState(context.Background())
			require.NoError(t, err)
			require.Equal(t, "public-read", state.ACL)
			require.Equal(t, status, state.VersioningStatus)
		})
	}
}

func TestBucketState_RejectsMalformedResponsesAndPreservesServiceErrors(t *testing.T) {
	for _, status := range []int{200, 403} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			client := localOSSClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(status)
				body := `<Unexpected/>`
				if status == 403 {
					body = `<Error><Code>AccessDenied</Code><Message>denied</Message><RequestId>0123456789ABCDEF01234567</RequestId></Error>`
				}
				_, err := fmt.Fprint(w, body)
				require.NoError(t, err)
			})
			state, err := client.BucketState(context.Background())
			require.Error(t, err)
			require.Nil(t, state)
			if status == 403 {
				var original *aliyunoss.ServiceError
				require.True(t, errors.As(err, &original))
				require.Equal(t, 403, original.StatusCode)
			}
		})
	}
}

func TestBucketState_RejectsUnexpectedVersionRootAndOversize(t *testing.T) {
	for _, body := range []string{`<Unexpected/>`, `<VersioningConfiguration>` + strings.Repeat(" ", 64*1024) + `</VersioningConfiguration>`} {
		client := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
			reply := body
			if r.URL.Query().Has("acl") {
				reply = `<AccessControlPolicy><AccessControlList><Grant>private</Grant></AccessControlList></AccessControlPolicy>`
			}
			_, err := fmt.Fprint(w, reply)
			require.NoError(t, err)
		})
		_, err := client.BucketState(context.Background())
		require.Error(t, err)
	}
}
