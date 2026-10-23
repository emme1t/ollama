package transfer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInitUploadResolvesRelativeLocation(t *testing.T) {
	for _, header := range []string{"Location", "Docker-Upload-Location"} {
		for _, tt := range []struct{ name, location, want string }{
			{"path relative", "session?state=token", "/prefix/v2/library/test/blobs/uploads/session?state=token"},
			{"query relative", "?state=token", "/prefix/v2/library/test/blobs/uploads/?state=token"},
			{"parent relative", "../session?state=token", "/prefix/v2/library/test/blobs/session?state=token"},
			{"root relative", "/session?state=token", "/session?state=token"},
		} {
			t.Run(header+"/"+tt.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/prefix/v2/library/test/blobs/uploads/" {
						t.Errorf("unexpected request path %q", r.URL.Path)
					}
					w.Header().Set(header, tt.location)
					w.WriteHeader(http.StatusAccepted)
				}))
				defer server.Close()
				u := &uploader{client: server.Client(), baseURL: server.URL + "/prefix", repository: "library/test"}
				ep, err := u.initUpload(t.Context(), Blob{Digest: "sha256:" + strings.Repeat("a", 64), Size: 1})
				if err != nil {
					t.Fatal(err)
				}
				if want := server.URL + tt.want; ep.sessionURL != want {
					t.Fatalf("session URL = %q, want %q", ep.sessionURL, want)
				}
			})
		}
	}
}
