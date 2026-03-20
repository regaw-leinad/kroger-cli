package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/regaw-leinad/kroger-cli/internal/secrets"
)

const (
	authorizeURL = "https://api.kroger.com/v1/connect/oauth2/authorize"
	tokenURL     = "https://api.kroger.com/v1/connect/oauth2/token"
	callbackPort = "8642"
	callbackPath = "/callback"
	scopes       = "product.compact cart.basic:write profile.compact"
)

// RedirectURI is the OAuth callback URI that must be registered in the Kroger developer portal.
const RedirectURI = "http://127.0.0.1:" + callbackPort + callbackPath

// Login performs the OAuth2 authorization code flow via a localhost callback.
func Login(ctx context.Context, store *secrets.Store, clientID, clientSecret string) error {
	state, err := randomState()
	if err != nil {
		return fmt.Errorf("generating state: %w", err)
	}

	authURL := buildAuthorizeURL(clientID, state)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	listener, err := net.Listen("tcp", "127.0.0.1:"+callbackPort)
	if err != nil {
		return fmt.Errorf("starting callback server: %w", err)
	}

	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
	}
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != callbackPath {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			errCh <- fmt.Errorf("state mismatch (possible CSRF)")
			return
		}
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			desc := r.URL.Query().Get("error_description")
			http.Error(w, "Authorization failed: "+desc, http.StatusBadRequest)
			errCh <- fmt.Errorf("authorization denied: %s - %s", errMsg, desc)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			errCh <- fmt.Errorf("no authorization code in callback")
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h2>Success!</h2><p>You can close this tab and return to the terminal.</p></body></html>`)
		codeCh <- code
	})

	go func() {
		if srvErr := srv.Serve(listener); srvErr != nil && srvErr != http.ErrServerClosed {
			errCh <- srvErr
		}
	}()
	defer srv.Close()

	if err := openBrowser(authURL); err != nil {
		return fmt.Errorf("opening browser: %w\n\nOpen this URL manually:\n%s", err, authURL)
	}

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}

	tok, err := exchangeCode(clientID, clientSecret, code)
	if err != nil {
		return fmt.Errorf("exchanging code for token: %w", err)
	}

	return store.SetToken(tok)
}

func buildAuthorizeURL(clientID, state string) string {
	v := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {RedirectURI},
		"response_type": {"code"},
		"scope":         {scopes},
		"state":         {state},
	}
	return authorizeURL + "?" + v.Encode()
}

func exchangeCode(clientID, clientSecret, code string) (secrets.Token, error) {
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {RedirectURI},
	}
	return requestToken(clientID, clientSecret, data)
}

// RefreshToken uses a refresh token to obtain new tokens.
func RefreshToken(clientID, clientSecret, refreshToken string) (secrets.Token, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	return requestToken(clientID, clientSecret, data)
}

func requestToken(clientID, clientSecret string, data url.Values) (secrets.Token, error) {
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return secrets.Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return secrets.Token{}, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return secrets.Token{}, fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return secrets.Token{}, fmt.Errorf("parsing token response: %w", err)
	}

	return secrets.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openBrowser(rawURL string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", rawURL).Start()
	case "linux":
		return exec.Command("xdg-open", rawURL).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL).Start()
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
}
