package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClientStreamPreservesHTTPStatusOnTruncatedErrorBody(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusBadGateway} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			const body = `{"error":"upstream unavailable`
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", fmt.Sprint(len(body)+100))
				w.WriteHeader(status)
				io.WriteString(w, body)
			}))
			defer server.Close()
			base, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			err = NewClient(base, server.Client()).Chat(t.Context(), &ChatRequest{Model: "test"}, func(ChatResponse) error { t.Error("unexpected callback"); return nil })
			var statusErr StatusError
			if !errors.As(err, &statusErr) || statusErr.StatusCode != status || statusErr.ErrorMessage != body {
				t.Fatalf("Chat error = %#v, want status %d and original body", err, status)
			}
		})
	}
}
