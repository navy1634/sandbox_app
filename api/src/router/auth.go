package router

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/sandbox-nextjs/src/config"
	"github.com/sandbox-nextjs/src/domain"
	"github.com/sandbox-nextjs/src/repository"
)

type authProxy struct {
	cfg      config.Config
	accounts repository.AccountRepository
	client   *http.Client
}

type authMeResponse struct {
	Authenticated     bool                   `json:"authenticated"`
	Account           map[string]interface{} `json:"account,omitempty"`
	NeedsRegistration bool                   `json:"needsRegistration,omitempty"`
	RedirectTo        string                 `json:"redirectTo,omitempty"`
	User              *authUser              `json:"user,omitempty"`
	AppAccount        *domain.AppAccount     `json:"appAccount,omitempty"`
}

type authUser struct {
	AccountID         int64  `json:"accountId"`
	Provider          string `json:"provider,omitempty"`
	ProviderAccountID string `json:"providerAccountId,omitempty"`
	Email             string `json:"email,omitempty"`
	Name              string `json:"name,omitempty"`
	Picture           string `json:"picture,omitempty"`
}

func RegisterRoutes(engine *gin.Engine, cfg config.Config, accounts repository.AccountRepository) error {
	proxy := authProxy{cfg: cfg, accounts: accounts, client: http.DefaultClient}

	engine.Use(corsMiddleware(cfg.FrontendURL))
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	engine.GET("/me", proxy.Me)
	engine.POST("/auth/logout", proxy.Logout)

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

func (p authProxy) Me(c *gin.Context) {
	res, err := p.forward(c, http.MethodGet, "/me")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to connect auth service"})
		return
	}
	defer func() {
		_ = res.Body.Close()
	}()
	p.copyCookies(c, res)

	if res.StatusCode != http.StatusOK {
		p.copyResponse(c, res)
		return
	}

	var me authMeResponse
	if err := json.NewDecoder(res.Body).Decode(&me); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to decode auth response"})
		return
	}
	if me.Authenticated && me.User != nil {
		account, err := p.accounts.UpsertFromAuthUser(c.Request.Context(), domain.AuthUser{
			AccountID: me.User.AccountID,
			Email:     me.User.Email,
			Name:      me.User.Name,
			Picture:   me.User.Picture,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save app account"})
			return
		}
		me.AppAccount = &account
	}

	c.JSON(http.StatusOK, me)
}

func (p authProxy) Logout(c *gin.Context) {
	res, err := p.forward(c, http.MethodPost, "/auth/logout")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to connect auth service"})
		return
	}
	defer func() {
		_ = res.Body.Close()
	}()
	p.copyCookies(c, res)
	p.copyResponse(c, res)
}

func (p authProxy) forward(c *gin.Context, method string, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.Request.Context(), method, p.serverURL(path).String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", c.GetHeader("Cookie"))
	req.Header.Set("Origin", c.GetHeader("Origin"))
	return p.client.Do(req)
}

func (p authProxy) copyCookies(c *gin.Context, res *http.Response) {
	for _, cookie := range res.Header.Values("Set-Cookie") {
		c.Writer.Header().Add("Set-Cookie", cookie)
	}
}

func (p authProxy) copyResponse(c *gin.Context, res *http.Response) {
	c.Header("Content-Type", res.Header.Get("Content-Type"))
	c.Status(res.StatusCode)
	if _, err := io.Copy(c.Writer, res.Body); err != nil {
		_ = c.Error(err)
	}
}

func (p authProxy) serverURL(path string) *url.URL {
	base, _ := url.Parse(p.cfg.AuthServerURL)
	return base.ResolveReference(&url.URL{Path: path})
}
