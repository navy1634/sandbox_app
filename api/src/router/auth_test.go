package router

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sandbox-nextjs/src/config"
	"github.com/sandbox-nextjs/src/infrastructure/session"
)

func TestLoginRedirectsToOIDCProviderWithRegisteredCallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var provider *httptest.Server
	provider = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 provider.URL,
			"authorization_endpoint": provider.URL + "/authorize",
			"token_endpoint":         provider.URL + "/token",
			"jwks_uri":               provider.URL + "/jwks",
		})
	}))
	defer provider.Close()

	cfg := config.Config{
		FrontendURL:      "https://app.example.com",
		OIDCIssuerURL:    provider.URL,
		OIDCInternalURL:  provider.URL,
		OIDCClientID:     "external-app",
		OIDCClientSecret: "client-secret",
		OIDCRedirectURL:  "https://app.example.com/auth/callback",
		SessionSecret:    []byte("01234567890123456789012345678901"),
	}
	engine := gin.New()
	if err := RegisterRoutes(engine, cfg, nil); err != nil {
		t.Fatalf("RegisterRoutes() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if location.Query().Get("redirect_uri") != cfg.OIDCRedirectURL {
		t.Fatalf("redirect_uri = %q, want %q", location.Query().Get("redirect_uri"), cfg.OIDCRedirectURL)
	}
	if location.Query().Get("redirect_to") != "" {
		t.Fatal("unexpected legacy redirect_to parameter")
	}
	if location.Query().Get("code_challenge_method") != "S256" {
		t.Fatalf("code_challenge_method = %q, want S256", location.Query().Get("code_challenge_method"))
	}
}

func TestCallbackExchangesCodeAndIssuesAppSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	var provider *httptest.Server
	nonce := ""
	provider = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 provider.URL,
				"authorization_endpoint": provider.URL + "/authorize",
				"token_endpoint":         provider.URL + "/token",
				"jwks_uri":               provider.URL + "/jwks",
			})
		case "/jwks":
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
				"kty": "RSA",
				"alg": "RS256",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1}),
			}}})
		case "/token":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			idToken := signRouterTestJWT(t, key, provider.URL, nonce)
			_ = json.NewEncoder(w).Encode(map[string]string{"id_token": idToken})
		default:
			http.NotFound(w, r)
		}
	}))
	defer provider.Close()

	cfg := config.Config{
		FrontendURL:      "https://app.example.com",
		OIDCIssuerURL:    provider.URL,
		OIDCInternalURL:  provider.URL,
		OIDCClientID:     "external-app",
		OIDCClientSecret: "client-secret",
		OIDCRedirectURL:  "https://app.example.com/auth/callback",
		SessionSecret:    []byte("01234567890123456789012345678901"),
	}
	engine := gin.New()
	if err := RegisterRoutes(engine, cfg, nil); err != nil {
		t.Fatalf("RegisterRoutes() error = %v", err)
	}

	loginRequest := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	loginResponse := httptest.NewRecorder()
	engine.ServeHTTP(loginResponse, loginRequest)
	loginLocation, err := url.Parse(loginResponse.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse login location: %v", err)
	}
	nonce = loginLocation.Query().Get("nonce")
	state := loginLocation.Query().Get("state")
	transactionCookie := cookieByName(loginResponse.Result().Cookies(), transactionCookieName)
	if nonce == "" || state == "" || transactionCookie == nil {
		t.Fatalf("login response did not contain OIDC transaction: nonce=%q state=%q cookie=%v", nonce, state, transactionCookie)
	}

	callbackRequest := httptest.NewRequest(http.MethodGet, "/auth/callback?code=authorization-code&state="+url.QueryEscape(state), nil)
	callbackRequest.AddCookie(transactionCookie)
	callbackResponse := httptest.NewRecorder()
	engine.ServeHTTP(callbackResponse, callbackRequest)
	if callbackResponse.Code != http.StatusFound {
		t.Fatalf("callback status = %d, want %d", callbackResponse.Code, http.StatusFound)
	}
	if callbackResponse.Header().Get("Location") != "https://app.example.com/dashboard" {
		t.Fatalf("callback location = %q", callbackResponse.Header().Get("Location"))
	}
	appCookie := cookieByName(callbackResponse.Result().Cookies(), sessionCookieName)
	if appCookie == nil || !appCookie.HttpOnly || !appCookie.Secure {
		t.Fatalf("callback did not issue a secure HttpOnly app session cookie: %v", appCookie)
	}
	user, err := session.NewManager(cfg.SessionSecret).Verify(appCookie.Value)
	if err != nil {
		t.Fatalf("verify app session: %v", err)
	}
	if user.Subject != "subject" || user.Email != "user@example.com" || !user.EmailVerified {
		t.Fatalf("unexpected app session user: %+v", user)
	}
}

func signRouterTestJWT(t *testing.T, key *rsa.PrivateKey, issuer string, nonce string) string {
	t.Helper()
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "test-key", "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{
		"iss":            issuer,
		"sub":            "subject",
		"aud":            "external-app",
		"exp":            time.Now().Add(time.Minute).Unix(),
		"nonce":          nonce,
		"email":          "user@example.com",
		"email_verified": true,
	})
	headerPart := base64.RawURLEncoding.EncodeToString(header)
	claimsPart := base64.RawURLEncoding.EncodeToString(claims)
	message := headerPart + "." + claimsPart
	digest := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return message + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func cookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
