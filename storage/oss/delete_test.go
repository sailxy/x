package oss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteObjectsUsesProvidedContextAndReturnsDeletedKeys(t *testing.T) {
	calls := 0
	client := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		require.Equal(t, http.MethodPost, r.Method)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "a.txt")
		require.Contains(t, string(body), "b.txt")
		require.NotContains(t, string(body), "<Quiet>true</Quiet>")
		_, err = fmt.Fprint(w, `<DeleteResult><Deleted><Key>path/a.txt</Key></Deleted></DeleteResult>`)
		require.NoError(t, err)
	})
	keys, err := client.DeleteObjects(context.Background(), []string{"path/a.txt", "path/b.txt"})
	require.NoError(t, err)
	require.Equal(t, []string{"path/a.txt"}, keys)
	require.Equal(t, 1, calls)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.DeleteObjects(ctx, []string{"path/a.txt"})
	require.ErrorIs(t, err, context.Canceled)
}
func TestDeleteObjectsSkipsSDKCallWithoutKeys(t *testing.T) {
	keys, err := (&Client{}).DeleteObjects(context.Background(), nil)
	require.NoError(t, err)
	require.Nil(t, keys)
}
func TestDeleteObjectUsesProvidedContext(t *testing.T) {
	client := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.True(t, strings.HasSuffix(r.URL.Path, "/path/a.txt"))
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, client.DeleteObject(context.Background(), "path/a.txt"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, client.DeleteObject(ctx, "path/a.txt"), context.Canceled)
}
