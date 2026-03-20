package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds non-sensitive CLI configuration.
type Config struct {
	DefaultStoreID    string `json:"default_store_id,omitempty"`
	DefaultStoreName  string `json:"default_store_name,omitempty"`
	DefaultStoreChain string `json:"default_store_chain,omitempty"`
	path              string
}

// ChainDomain maps a Kroger chain code to its website domain.
var ChainDomain = map[string]string{
	"FRED":          "www.fredmeyer.com",
	"KROGER":        "www.kroger.com",
	"RALPHS":        "www.ralphs.com",
	"HARRIS TEETER": "www.harristeeter.com",
	"KING SOOPERS":  "www.kingsoopers.com",
	"DILLONS":       "www.dillons.com",
	"FRYS":          "www.frysfood.com",
	"QFC":           "www.qfc.com",
	"SMITHS":        "www.smithsfoodanddrug.com",
	"FOOD 4 LESS":   "www.food4less.com",
	"MARIANOS":      "www.marianos.com",
	"PICK N SAVE":   "www.picknsave.com",
	"METRO MARKET":  "www.metromarket.net",
	"CITY MARKET":   "www.citymarket.com",
}

// StoreDomain returns the website domain for the configured store chain.
// Defaults to www.kroger.com if the chain is unknown.
func (c *Config) StoreDomain() string {
	if d, ok := ChainDomain[c.DefaultStoreChain]; ok {
		return d
	}
	return "www.kroger.com"
}

// Dir returns the configuration directory path.
func Dir() string {
	if d := os.Getenv("KROGER_CONFIG_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".config", "kroger")
	}
	return filepath.Join(home, ".config", "kroger")
}

// Path returns the config file path.
func Path() string {
	return filepath.Join(Dir(), "config.json")
}

// Default returns an empty config.
func Default() *Config {
	return &Config{path: Path()}
}

// Load reads the config from disk. Returns Default() if the file doesn't exist.
func Load() (*Config, error) {
	p := Path()
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &Config{path: p}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg.path = p
	return &cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(c.path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// SetStore saves the default store.
func (c *Config) SetStore(id, name, chain string) error {
	c.DefaultStoreID = id
	c.DefaultStoreName = name
	c.DefaultStoreChain = chain
	return c.Save()
}
