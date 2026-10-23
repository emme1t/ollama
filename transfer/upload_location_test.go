package transfer

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestInitUploadLocation(t *testing.T) {
	for _, header := range []string{"Location", "Docker-Upload-Location"} {
		for _, tt := range []struct {
			name, location string
			wantErr        bool
		}{
			{"invalid escape", "/upload/%zz", true},
			{"invalid port", "http://registry.example:bad/upload", true},
			{"relative", "/upload/session?state=token", false},
		} {
			t.Run(header+"/"+tt.name, func(t *testing.T) {
				defer func() {
					if p := recover(); p != nil {
						t.Fatalf("initUpload panicked instead of returning a URL error: %v", p)
					}
				}()
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set(header, tt.location)
					w.WriteHeader(http.StatusAccepted)
				}))
				defer server.Close()
				u := &uploader{client: server.Client(), baseURL: server.URL, repository: "library/test"}
				ep, err := u.initUpload(t.Context(), Blob{Digest: "sha256:" + strings.Repeat("a", 64), Size: 1})
				if tt.wantErr {
					var urlErr *url.Error
					if !errors.As(err, &urlErr) {
						t.Fatalf("initUpload error = %v, want URL parse error", err)
					}
					if ep.sessionURL != "" {
						t.Fatalf("invalid location produced session URL %q", ep.sessionURL)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if want := server.URL + tt.location; ep.sessionURL != want {
					t.Fatalf("session URL = %q, want %q", ep.sessionURL, want)
				}
			})
		}
	}
}
