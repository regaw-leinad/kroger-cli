package secrets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Credentials holds the Kroger app credentials.
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// Token holds OAuth2 tokens.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// IsExpired returns true if the access token has expired (with 60s buffer).
func (t Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt.Add(-60 * time.Second))
}

type storeData struct {
	Credentials *Credentials `json:"credentials,omitempty"`
	Token       *Token       `json:"token,omitempty"`
}

// Store provides access to secrets stored in a local file.
type Store struct {
	path string
}

// Open creates a new Store backed by a local file.
func Open() (*Store, error) {
	dir := dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("creating secrets dir: %w", err)
	}
	return &Store{path: filepath.Join(dir, "credentials.json")}, nil
}

func dir() string {
	if d := os.Getenv("KROGER_CONFIG_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "kroger")
}

func (s *Store) load() (*storeData, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return &storeData{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading credentials file: %w", err)
	}
	var d storeData
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parsing credentials file: %w", err)
	}
	return &d, nil
}

func (s *Store) save(d *storeData) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling credentials: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0600)
}

// SetCredentials stores the app credentials.
func (s *Store) SetCredentials(creds Credentials) error {
	d, err := s.load()
	if err != nil {
		return err
	}
	d.Credentials = &creds
	return s.save(d)
}

// GetCredentials retrieves the app credentials.
func (s *Store) GetCredentials() (Credentials, error) {
	d, err := s.load()
	if err != nil {
		return Credentials{}, err
	}
	if d.Credentials == nil {
		return Credentials{}, nil
	}
	return *d.Credentials, nil
}

// SetToken stores the OAuth tokens.
func (s *Store) SetToken(tok Token) error {
	d, err := s.load()
	if err != nil {
		return err
	}
	d.Token = &tok
	return s.save(d)
}

// GetToken retrieves the OAuth tokens.
func (s *Store) GetToken() (Token, error) {
	d, err := s.load()
	if err != nil {
		return Token{}, err
	}
	if d.Token == nil {
		return Token{}, nil
	}
	return *d.Token, nil
}

// DeleteToken removes the stored OAuth tokens.
func (s *Store) DeleteToken() error {
	d, err := s.load()
	if err != nil {
		return err
	}
	d.Token = nil
	return s.save(d)
}

// DeleteCredentials removes the stored app credentials.
func (s *Store) DeleteCredentials() error {
	d, err := s.load()
	if err != nil {
		return err
	}
	d.Credentials = nil
	return s.save(d)
}
