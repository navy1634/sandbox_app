package router

import (
	"crypto/hmac"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sandbox-nextjs/src/config"
	"github.com/sandbox-nextjs/src/domain"
	"github.com/sandbox-nextjs/src/infrastructure/oidc"
	"github.com/sandbox-nextjs/src/infrastructure/session"
	"github.com/sandbox-nextjs/src/repository"
)

const (
	sessionCookieName     = "app_session"
	transactionCookieName = "oidc_transaction"
)

type authHandler struct {
	cfg      config.Config
	accounts repository.AccountRepository
	oidc     *oidc.Client
	sessions *session.Manager
}

type meResponse struct {
	Authenticated bool               `json:"authenticated"`
	User          *authUser          `json:"user,omitempty"`
	AppAccount    *domain.AppAccount `json:"appAccount,omitempty"`
}

type authUser struct {
	Subject       string `json:"sub"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"emailVerified,omitempty"`
}

func RegisterRoutes(engine *gin.Engine, cfg config.Config, accounts repository.AccountRepository) error {
	oidcClient, err := oidc.NewClient(oidc.Config{
		IssuerURL:    cfg.OIDCIssuerURL,
		InternalURL:  cfg.OIDCInternalURL,
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		RedirectURL:  cfg.OIDCRedirectURL,
	}, nil)
	if err != nil {
		return err
	}
	handler := authHandler{
		cfg:      cfg,
		accounts: accounts,
		oidc:     oidcClient,
		sessions: session.NewManager(cfg.SessionSecret),
	}

	engine.Use(corsMiddleware(cfg.FrontendURL))
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	engine.GET("/me", handler.Me)
	engine.GET("/auth/login", handler.Login)
	engine.GET("/auth/callback", handler.Callback)
	engine.POST("/auth/logout", handler.Logout)

	return nil
}

func corsMiddleware(frontendURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", frontendURL)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// OIDC認可を開始し、state・nonce・PKCE verifierを署名Cookieに保存する。
func (h authHandler) Login(c *gin.Context) {
	authorizationURL, request, err := h.oidc.AuthorizationURL(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to prepare OIDC authorization"})
		return
	}
	transaction, err := h.sessions.SignTransaction(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create OIDC transaction"})
		return
	}
	h.setCookie(c, transactionCookieName, transaction, int(session.TransactionTTL/time.Second), true, "/auth")
	c.Redirect(http.StatusFound, authorizationURL)
}

// OIDC callbackで認可コードを交換し、アプリ用セッションを発行する。
func (h authHandler) Callback(c *gin.Context) {
	transactionValue, err := c.Cookie(transactionCookieName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OIDC transaction is missing"})
		return
	}
	h.clearCookie(c, transactionCookieName, "/auth")
	transaction, err := h.sessions.VerifyTransaction(transactionValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OIDC transaction is invalid"})
		return
	}
	if providerError := c.Query("error"); providerError != "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OIDC authorization failed"})
		return
	}
	if !hmac.Equal([]byte(transaction.State), []byte(c.Query("state"))) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OIDC state is invalid"})
		return
	}
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OIDC authorization code is missing"})
		return
	}
	user, err := h.oidc.Exchange(c.Request.Context(), code, transaction)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OIDC token validation failed"})
		return
	}
	sessionValue, err := h.sessions.Sign(session.User{
		Subject:       user.Subject,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create app session"})
		return
	}
	h.setCookie(c, sessionCookieName, sessionValue, int(session.DefaultTTL/time.Second), true, "/")
	c.Redirect(http.StatusFound, h.dashboardURL())
}

// アプリのセッションからユーザー情報を読み、アカウントを保存する。
func (h authHandler) Me(c *gin.Context) {
	value, err := c.Cookie(sessionCookieName)
	if err != nil {
		c.JSON(http.StatusOK, meResponse{Authenticated: false})
		return
	}
	user, err := h.sessions.Verify(value)
	if err != nil {
		c.JSON(http.StatusOK, meResponse{Authenticated: false})
		return
	}
	account, err := h.accounts.UpsertFromOIDCUser(c.Request.Context(), domain.OIDCUser{
		Subject:       user.Subject,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save app account"})
		return
	}
	response := meResponse{
		Authenticated: true,
		User: &authUser{
			Subject:       user.Subject,
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
		},
	}
	response.AppAccount = &account
	c.JSON(http.StatusOK, response)
}

// アプリセッションだけを削除する。
func (h authHandler) Logout(c *gin.Context) {
	h.clearCookie(c, sessionCookieName, "/")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h authHandler) dashboardURL() string {
	frontendURL, err := url.Parse(h.cfg.FrontendURL)
	if err != nil {
		return "/dashboard"
	}
	frontendURL.Path = strings.TrimRight(frontendURL.Path, "/") + "/dashboard"
	frontendURL.RawPath = ""
	frontendURL.RawQuery = ""
	return frontendURL.String()
}

func (h authHandler) setCookie(c *gin.Context, name string, value string, maxAge int, httpOnly bool, path string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, path, "", strings.HasPrefix(h.cfg.FrontendURL, "https://"), httpOnly)
}

func (h authHandler) clearCookie(c *gin.Context, name string, path string) {
	h.setCookie(c, name, "", -1, true, path)
}
