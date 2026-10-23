package transfer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseAuthChallengeParameters(t *testing.T) {
	want := AuthChallenge{Realm: "https://auth.example/token", Service: "Registry", Scope: "repository:library/model:pull,push"}
	for _, tt := range []struct{ name, header string }{
		{"standard", `Bearer realm="https://auth.example/token",service="Registry",scope="repository:library/model:pull,push"`},
		{"case insensitive names", `bearer Realm="https://auth.example/token",SERVICE="Registry",Scope="repository:library/model:pull,push"`},
		{"whitespace around equals", "Bearer realm \t= \t\"https://auth.example/token\", service = Registry, scope = \"repository:library/model:pull,push\""},
		{"quoted extra parameter", `Bearer note="realm=wrong, service=wrong, scope=wrong",realm="https://auth.example/token",service=Registry,scope="repository:library/model:pull,push"`},
		{"quoted pairs", `Bearer realm="https:\/\/auth.example\/token",service="Regi\stry",scope="repository:library/model:pull,push"`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseAuthChallenge(tt.header); got != want {
				t.Fatalf("challenge = %#v, want %#v", got, want)
			}
		})
	}
}

func TestUploadExistsAuthChallengeParameters(t *testing.T) {
	want := AuthChallenge{Realm: "https://auth.example/token", Service: "Registry", Scope: "repository:library/model:pull,push"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer fixture-token" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("WWW-Authenticate", `bearer Realm = "https://auth.example/token", SERVICE = Registry, Scope = "repository:library/model:pull,push"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	refreshes := 0
	u := &uploader{
		client: server.Client(), baseURL: server.URL, repository: "library/model",
		getToken: func(_ context.Context, got AuthChallenge) (string, error) {
			refreshes++
			if got != want {
				return "", fmt.Errorf("challenge = %#v, want %#v", got, want)
			}
			return "fixture-token", nil
		},
	}
	exists, err := u.exists(t.Context(), Blob{Digest: "sha256:" + strings.Repeat("a", 64), Size: 1})
	if err != nil || !exists {
		t.Fatalf("exists = %v, error = %v", exists, err)
	}
	if refreshes != 1 {
		t.Fatalf("token callback called %d times, want 1", refreshes)
	}
}
