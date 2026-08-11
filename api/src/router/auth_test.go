package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sandbox-nextjs/src/config"
)

func TestAuthProxyForwardPreservesCookieAndOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Cookie"); got != "app_session=signed-session" {
			t.Errorf("Cookie = %q, want %q", got, "app_session=signed-session")
		}
		if got := r.Header.Get("Origin"); got != "https://app.sandbox.navy1634.com" {
			t.Errorf("Origin = %q, want %q", got, "https://app.sandbox.navy1634.com")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	proxy := authProxy{
		cfg:    config.Config{AuthServerURL: upstream.URL},
		client: upstream.Client(),
	}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	context.Request.Header.Set("Cookie", "app_session=signed-session")
	context.Request.Header.Set("Origin", "https://app.sandbox.navy1634.com")

	response, err := proxy.forward(context, http.MethodPost, "/auth/logout")
	if err != nil {
		t.Fatalf("forward() error = %v", err)
	}
	defer response.Body.Close()
}
