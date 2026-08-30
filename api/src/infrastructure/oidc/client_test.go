package oidc

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
)

func TestAuthorizationURLUsesDiscoveryAndPKCE(t *testing.T) {
	server := newOIDCTestServer(t, false)
	defer server.Close()

	client, err := NewClient(Config{
		IssuerURL:    server.URL,
		InternalURL:  server.URL,
		ClientID:     "client",
		ClientSecret: "secret",
		RedirectURL:  "https://app.example.com/auth/callback",
	}, server.Client())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	authorizationURL, request, err := client.AuthorizationURL(t.Context())
	if err != nil {
		t.Fatalf("AuthorizationURL() error = %v", err)
	}
	parsedURL, err := url.Parse(authorizationURL)
	if err != nil {
		t.Fatalf("parse authorization URL: %v", err)
	}
	query := parsedURL.Query()
	if query.Get("client_id") != "client" || query.Get("redirect_uri") != "https://app.example.com/auth/callback" {
		t.Fatalf("unexpected authorization query: %s", parsedURL.RawQuery)
	}
	if query.Get("response_type") != "code" || query.Get("code_challenge_method") != "S256" {
		t.Fatalf("unexpected authorization flow: %s", parsedURL.RawQuery)
	}
	if query.Get("state") != request.State || query.Get("nonce") != request.Nonce {
		t.Fatalf("authorization state does not match request: %+v", request)
	}
	if query.Get("code_challenge") != testCodeChallenge(request.CodeVerifier) {
		t.Fatalf("authorization code challenge does not match request")
	}
}

func TestExchangeValidatesIDTokenClaimsAndSignature(t *testing.T) {
	server := newOIDCTestServer(t, true)
	defer server.Close()

	client, err := NewClient(Config{
		IssuerURL:    server.URL,
		InternalURL:  server.URL,
		ClientID:     "client",
		ClientSecret: "secret",
		RedirectURL:  "https://app.example.com/auth/callback",
	}, server.Client())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	user, err := client.Exchange(t.Context(), "code", AuthorizationRequest{Nonce: "nonce", CodeVerifier: "verifier"})
	if err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}
	if user.Subject != "subject" || user.Email != "user@example.com" || !user.EmailVerified {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestExchangeRejectsNonceMismatch(t *testing.T) {
	server := newOIDCTestServer(t, true)
	defer server.Close()

	client, err := NewClient(Config{
		IssuerURL:    server.URL,
		InternalURL:  server.URL,
		ClientID:     "client",
		ClientSecret: "secret",
		RedirectURL:  "https://app.example.com/auth/callback",
	}, server.Client())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if _, err := client.Exchange(t.Context(), "code", AuthorizationRequest{Nonce: "wrong", CodeVerifier: "verifier"}); err == nil {
		t.Fatal("expected nonce mismatch to be rejected")
	}
}

func newOIDCTestServer(t *testing.T, tokenEnabled bool) *httptest.Server {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 server.URL,
				"authorization_endpoint": server.URL + "/authorize",
				"token_endpoint":         server.URL + "/token",
				"jwks_uri":               server.URL + "/jwks",
			})
		case "/jwks":
			publicKey := &key.PublicKey
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
				"kty": "RSA",
				"alg": "RS256",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1}),
			}}})
		case "/token":
			if !tokenEnabled {
				http.NotFound(w, r)
				return
			}
			claims := map[string]any{
				"iss":            server.URL,
				"sub":            "subject",
				"aud":            "client",
				"exp":            time.Now().Add(time.Minute).Unix(),
				"iat":            time.Now().Unix(),
				"nonce":          "nonce",
				"email":          "user@example.com",
				"email_verified": true,
			}
			idToken := signTestJWT(t, key, claims)
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "access-token", "token_type": "Bearer", "id_token": idToken})
		default:
			http.NotFound(w, r)
		}
	}))
	return server
}

func signTestJWT(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "test-key", "typ": "JWT"})
	payload, _ := json.Marshal(claims)
	headerPart := base64.RawURLEncoding.EncodeToString(header)
	payloadPart := base64.RawURLEncoding.EncodeToString(payload)
	message := headerPart + "." + payloadPart
	digest := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return message + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func testCodeChallenge(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}
