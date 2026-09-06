package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRESTGet_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected method GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	r := NewREST()
	defer func() { _ = r.Close() }()

	resp, err := r.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "ok", resp.String())
}

func TestRESTGet_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer ts.Close()

	r := NewREST()
	defer func() { _ = r.Close() }()

	resp, err := r.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
	assert.Equal(t, "not found", resp.String())
}

func TestNewRESTWithClient(t *testing.T) {
	t.Run("default timeout", func(t *testing.T) {
		r := NewREST()
		defer func() { _ = r.Close() }()
		assert.Equal(t, DefaultTimeout, r.doTimeout)
		assert.Zero(t, r.client.Client().Timeout)
	})

	t.Run("preserve supplied client", func(t *testing.T) {
		supplied := &http.Client{Timeout: time.Second}
		r := NewRESTWithClient(supplied)
		defer func() { _ = r.Close() }()
		assert.Same(t, supplied, r.client.Client())
		assert.Equal(t, DefaultTimeout, r.doTimeout)
	})
}

func TestRESTDo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("X-REST-Test", "yes")
			_, _ = io.WriteString(w, `{"ok":true}`)
		}))
		defer server.Close()

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		r := NewRESTWithClient(server.Client())
		defer func() { _ = r.Close() }()
		resp, err := r.Do(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "yes", resp.Header.Get("X-REST-Test"))
		assert.JSONEq(t, `{"ok":true}`, string(resp.Body))
	})

	t.Run("non-2xx returns bounded response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			_, _ = io.WriteString(w, `{"error":"rejected"}`)
		}))
		defer server.Close()

		req, err := http.NewRequest(http.MethodPost, server.URL, nil)
		require.NoError(t, err)
		r := NewRESTWithClient(server.Client())
		defer func() { _ = r.Close() }()
		resp, err := r.Do(context.Background(), req)

		var statusErr *StatusError
		require.ErrorAs(t, err, &statusErr)
		assert.Equal(t, http.StatusTeapot, statusErr.StatusCode)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusTeapot, resp.StatusCode)
		assert.JSONEq(t, `{"error":"rejected"}`, string(resp.Body))
	})

	t.Run("response body limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, strings.Repeat("x", int(MaxResponseBodyBytes)+1))
		}))
		defer server.Close()

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		r := NewRESTWithClient(server.Client())
		defer func() { _ = r.Close() }()
		resp, err := r.Do(context.Background(), req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrResponseTooLarge)
	})

	t.Run("canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		req, err := http.NewRequest(http.MethodGet, "https://platform.invalid", nil)
		require.NoError(t, err)
		r := NewRESTWithClient(&http.Client{Transport: waitForContextTransport{}})
		defer func() { _ = r.Close() }()
		resp, err := r.Do(ctx, req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("context deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		req, err := http.NewRequest(http.MethodGet, "https://platform.invalid", nil)
		require.NoError(t, err)
		r := NewRESTWithClient(&http.Client{Transport: waitForContextTransport{}})
		defer func() { _ = r.Close() }()
		resp, err := r.Do(ctx, req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("invalid inputs", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "https://platform.invalid", nil)
		require.NoError(t, err)
		r := NewREST()
		defer func() { _ = r.Close() }()
		_, err = r.Do(nil, req)
		assert.EqualError(t, err, "REST context is nil")
		_, err = r.Do(context.Background(), nil)
		assert.EqualError(t, err, "REST request is nil")
	})
}

type waitForContextTransport struct{}

func (waitForContextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}

func TestStatusErrorDoesNotContainBody(t *testing.T) {
	err := &StatusError{StatusCode: http.StatusUnauthorized}
	assert.Equal(t, "REST request returned status 401", err.Error())
	assert.False(t, errors.Is(err, ErrResponseTooLarge))
}

func TestResponseFormattingOmitsHeadersAndBody(t *testing.T) {
	resp := Response{
		StatusCode: http.StatusUnauthorized,
		Header:     http.Header{"Authorization": {"Bearer response-secret"}},
		Body:       []byte(`{"access_token":"response-secret"}`),
	}

	for _, format := range []string{"%s", "%q", "%v", "%+v", "%#v"} {
		formatted := fmt.Sprintf(format, resp)
		assert.NotContains(t, formatted, "response-secret")
		assert.Contains(t, formatted, "401")
	}
}
