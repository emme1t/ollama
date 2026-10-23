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

func TestClientChatTruncatedHTTPStream(t *testing.T) {
	for _, tt := range []struct {
		name       string
		body       string
		truncated  bool
		wantChunks int
	}{
		{"partial first chunk", `{"message":{"content":"hel`, true, 0},
		{"partial later chunk", "{\"message\":{\"content\":\"hello\"}}\n" + `{"message":{"content":"wor`, true, 1},
		{"malformed complete response", `{"message":{"content":"hel`, false, 0},
		{"valid final line without newline", `{"message":{"content":"hello"},"done":true}`, false, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/x-ndjson")
				size := len(tt.body)
				if tt.truncated {
					size += 100
				}
				w.Header().Set("Content-Length", fmt.Sprint(size))
				io.WriteString(w, tt.body)
			}))
			defer ts.Close()
			base, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}
			client := NewClient(base, ts.Client())
			chunks := 0
			err = client.Chat(t.Context(), &ChatRequest{Model: "test"}, func(ChatResponse) error {
				chunks++
				return nil
			})
			if tt.truncated {
				if !errors.Is(err, io.ErrUnexpectedEOF) {
					t.Errorf("Chat error = %v, want unexpected EOF", err)
				}
			} else if tt.wantChunks == 0 {
				if err == nil || errors.Is(err, io.ErrUnexpectedEOF) {
					t.Errorf("Chat error = %v, want JSON error", err)
				}
			} else if err != nil {
				t.Errorf("Chat error = %v", err)
			}
			if chunks != tt.wantChunks {
				t.Errorf("callback count = %d, want %d", chunks, tt.wantChunks)
			}
		})
	}
}
