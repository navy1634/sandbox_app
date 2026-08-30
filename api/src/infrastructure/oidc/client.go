package oidc

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Config struct {
	IssuerURL    string
	InternalURL  string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type AuthorizationRequest struct {
	State        string `json:"state"`
	Nonce        string `json:"nonce"`
	CodeVerifier string `json:"code_verifier"`
}

type User struct {
	Subject       string
	Email         string
	EmailVerified bool
}

type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	metadata   *metadata
	keys       map[string]*rsa.PublicKey
}

type metadata struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

type tokenResponse struct {
	IDToken string `json:"id_token"`
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyType   string `json:"kty"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

// OIDCクライアントを初期化する。
func NewClient(cfg Config, httpClient *http.Client) (*Client, error) {
	cfg.IssuerURL = strings.TrimRight(strings.TrimSpace(cfg.IssuerURL), "/")
	cfg.InternalURL = strings.TrimRight(strings.TrimSpace(cfg.InternalURL), "/")
	if cfg.InternalURL == "" {
		cfg.InternalURL = cfg.IssuerURL
	}
	if !validURL(cfg.IssuerURL) || !validURL(cfg.InternalURL) || !validURL(cfg.RedirectURL) {
		return nil, errors.New("OIDC URLs must be absolute http or https URLs")
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, errors.New("OIDC client credentials are required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{cfg: cfg, httpClient: httpClient}, nil
}

// DiscoveryからPKCE付きのOIDC認可URLを作る。
func (c *Client) AuthorizationURL(ctx context.Context) (string, AuthorizationRequest, error) {
	provider, err := c.provider(ctx)
	if err != nil {
		return "", AuthorizationRequest{}, err
	}
	request, err := newAuthorizationRequest()
	if err != nil {
		return "", AuthorizationRequest{}, err
	}
	authorizationURL, err := url.Parse(provider.AuthorizationEndpoint)
	if err != nil {
		return "", AuthorizationRequest{}, errors.New("OIDC authorization endpoint is invalid")
	}
	query := authorizationURL.Query()
	query.Set("client_id", c.cfg.ClientID)
	query.Set("redirect_uri", c.cfg.RedirectURL)
	query.Set("response_type", "code")
	query.Set("scope", "openid email profile")
	query.Set("state", request.State)
	query.Set("nonce", request.Nonce)
	query.Set("code_challenge", codeChallenge(request.CodeVerifier))
	query.Set("code_challenge_method", "S256")
	authorizationURL.RawQuery = query.Encode()
	return authorizationURL.String(), request, nil
}

// 認可コードを交換し、IDトークンを検証済みのユーザーへ変換する。
func (c *Client) Exchange(ctx context.Context, code string, request AuthorizationRequest) (User, error) {
	if code == "" || request.Nonce == "" || request.CodeVerifier == "" {
		return User{}, errors.New("OIDC authorization response is incomplete")
	}
	provider, err := c.provider(ctx)
	if err != nil {
		return User{}, err
	}
	tokenURL, err := c.internalEndpoint(provider.TokenEndpoint)
	if err != nil {
		return User{}, err
	}
	form := url.Values{
		"grant_type":    []string{"authorization_code"},
		"code":          []string{code},
		"redirect_uri":  []string{c.cfg.RedirectURL},
		"code_verifier": []string{request.CodeVerifier},
	}
	tokenRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return User{}, err
	}
	tokenRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenRequest.SetBasicAuth(c.cfg.ClientID, c.cfg.ClientSecret)
	response, err := c.httpClient.Do(tokenRequest)
	if err != nil {
		return User{}, err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		return User{}, errors.New("OIDC token exchange failed")
	}
	var tokens tokenResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&tokens); err != nil {
		return User{}, errors.New("OIDC token response is invalid")
	}
	if tokens.IDToken == "" {
		return User{}, errors.New("OIDC ID token is missing")
	}
	return c.verifyIDToken(ctx, provider, tokens.IDToken, request.Nonce)
}

func (c *Client) provider(ctx context.Context) (metadata, error) {
	c.mu.Lock()
	if c.metadata != nil {
		value := *c.metadata
		c.mu.Unlock()
		return value, nil
	}
	c.mu.Unlock()

	discoveryURL, err := url.Parse(c.cfg.InternalURL)
	if err != nil {
		return metadata{}, errors.New("OIDC internal URL is invalid")
	}
	discoveryURL.Path = strings.TrimRight(discoveryURL.Path, "/") + "/.well-known/openid-configuration"
	discoveryURL.RawPath = ""
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL.String(), nil)
	if err != nil {
		return metadata{}, err
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return metadata{}, err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		return metadata{}, errors.New("OIDC discovery failed")
	}
	var discovered metadata
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&discovered); err != nil {
		return metadata{}, errors.New("OIDC discovery response is invalid")
	}
	if discovered.Issuer != c.cfg.IssuerURL || !validURL(discovered.AuthorizationEndpoint) || !validURL(discovered.TokenEndpoint) || !validURL(discovered.JWKSURI) {
		return metadata{}, errors.New("OIDC discovery issuer or endpoint is invalid")
	}
	c.mu.Lock()
	c.metadata = &discovered
	c.mu.Unlock()
	return discovered, nil
}

func (c *Client) verifyIDToken(ctx context.Context, provider metadata, rawToken string, nonce string) (User, error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return User{}, errors.New("OIDC ID token format is invalid")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return User{}, errors.New("OIDC ID token header is invalid")
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Algorithm != "RS256" || header.KeyID == "" {
		return User{}, errors.New("OIDC ID token algorithm is invalid")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return User{}, errors.New("OIDC ID token payload is invalid")
	}
	var claims map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(payloadBytes)))
	decoder.UseNumber()
	if err := decoder.Decode(&claims); err != nil {
		return User{}, errors.New("OIDC ID token claims are invalid")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return User{}, errors.New("OIDC ID token signature is invalid")
	}
	publicKey, err := c.publicKey(ctx, provider, header.KeyID)
	if err != nil {
		return User{}, err
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature); err != nil {
		return User{}, errors.New("OIDC ID token signature is invalid")
	}
	if claims["iss"] != provider.Issuer || claims["nonce"] != nonce {
		return User{}, errors.New("OIDC ID token issuer or nonce is invalid")
	}
	if !audienceContains(claims["aud"], c.cfg.ClientID) {
		return User{}, errors.New("OIDC ID token audience is invalid")
	}
	expiresAt, ok := numericClaim(claims["exp"])
	if !ok || expiresAt <= time.Now().Unix() {
		return User{}, errors.New("OIDC ID token is expired")
	}
	subject, ok := claims["sub"].(string)
	if !ok || subject == "" {
		return User{}, errors.New("OIDC ID token subject is missing")
	}
	user := User{Subject: subject}
	if email, ok := claims["email"].(string); ok {
		user.Email = email
	}
	if emailVerified, ok := claims["email_verified"].(bool); ok {
		user.EmailVerified = emailVerified
	}
	return user, nil
}

func (c *Client) publicKey(ctx context.Context, provider metadata, keyID string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	if key, ok := c.keys[keyID]; ok {
		c.mu.Unlock()
		return key, nil
	}
	c.mu.Unlock()

	jwksURL, err := c.internalEndpoint(provider.JWKSURI)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("OIDC JWKS request failed")
	}
	var body jwksResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&body); err != nil {
		return nil, errors.New("OIDC JWKS response is invalid")
	}
	keys := map[string]*rsa.PublicKey{}
	for _, key := range body.Keys {
		if key.KeyType != "RSA" || key.Algorithm != "RS256" || key.KeyID == "" {
			continue
		}
		publicKey, err := parseJWK(key)
		if err != nil {
			return nil, err
		}
		keys[key.KeyID] = publicKey
	}
	c.mu.Lock()
	c.keys = keys
	publicKey := c.keys[keyID]
	c.mu.Unlock()
	if publicKey == nil {
		return nil, errors.New("OIDC signing key is not found")
	}
	return publicKey, nil
}

func (c *Client) internalEndpoint(publicEndpoint string) (string, error) {
	publicBase, err := url.Parse(c.cfg.IssuerURL)
	if err != nil {
		return "", err
	}
	internalBase, err := url.Parse(c.cfg.InternalURL)
	if err != nil {
		return "", err
	}
	endpoint, err := url.Parse(publicEndpoint)
	if err != nil || endpoint.Host == "" {
		return "", errors.New("OIDC endpoint is invalid")
	}
	issuerPath := strings.TrimRight(publicBase.Path, "/")
	if !strings.HasPrefix(endpoint.Path, issuerPath) {
		return "", errors.New("OIDC endpoint is outside issuer")
	}
	suffix := strings.TrimPrefix(endpoint.Path, issuerPath)
	internalBase.Path = strings.TrimRight(internalBase.Path, "/") + suffix
	internalBase.RawPath = ""
	internalBase.RawQuery = endpoint.RawQuery
	return internalBase.String(), nil
}

func parseJWK(key jwk) (*rsa.PublicKey, error) {
	modulus, err := base64.RawURLEncoding.DecodeString(key.Modulus)
	if err != nil || len(modulus) == 0 {
		return nil, errors.New("OIDC JWK modulus is invalid")
	}
	exponent, err := base64.RawURLEncoding.DecodeString(key.Exponent)
	if err != nil || len(exponent) == 0 {
		return nil, errors.New("OIDC JWK exponent is invalid")
	}
	exponentValue := 0
	for _, value := range exponent {
		exponentValue = exponentValue<<8 | int(value)
	}
	if exponentValue < 2 {
		return nil, errors.New("OIDC JWK exponent is invalid")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(modulus), E: exponentValue}, nil
}

func newAuthorizationRequest() (AuthorizationRequest, error) {
	state, err := randomString(32)
	if err != nil {
		return AuthorizationRequest{}, err
	}
	nonce, err := randomString(32)
	if err != nil {
		return AuthorizationRequest{}, err
	}
	verifier, err := randomString(32)
	if err != nil {
		return AuthorizationRequest{}, err
	}
	return AuthorizationRequest{State: state, Nonce: nonce, CodeVerifier: verifier}, nil
}

func randomString(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func codeChallenge(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func audienceContains(value any, expected string) bool {
	switch audience := value.(type) {
	case string:
		return audience == expected
	case []any:
		for _, item := range audience {
			if item == expected {
				return true
			}
		}
	}
	return false
}

func numericClaim(value any) (int64, bool) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	parsed, err := number.Int64()
	return parsed, err == nil
}

func validURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.IsAbs() && parsed.Host != "" && parsed.User == nil && parsed.Fragment == "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}
