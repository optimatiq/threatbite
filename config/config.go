package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config configuration struct for the project
type Config struct {
	Port              int
	Debug             bool
	PwnedKey          string
	MaxmindKey        string
	SMTPHello         string
	SMTPFrom          string
	AutoTLS           bool
	ProxyList         []string
	SpamList          []string
	VPNList           []string
	DCList            []string
	EmailDisposalList []string
	EmailFreeList     []string
}

// NewConfig returns a new configuration struct or error.
// Configuration is stored in the environment as the one of the tenets of a twelve-factor app.
// config.env or config_local.env files are allowed but are totally optional.
// If config_local.env is present than it's used (only this file), otherwise config.env is checked
// Envs take precedence of envs that are imported from config_local.env or config.env files
func NewConfig(configFile string) (*Config, error) {
	config := &Config{
		Port:              8080,
		Debug:             false,
		ProxyList:         []string{"https://get.threatbite.com/public/proxy.txt"},
		SpamList:          []string{"https://get.threatbite.com/public/spam.txt"},
		VPNList:           []string{"https://get.threatbite.com/public/vpn.txt"},
		DCList:            []string{"https://get.threatbite.com/public/dc-names.txt"},
		EmailDisposalList: []string{"https://get.threatbite.com/public/disposal.txt"},
		EmailFreeList:     []string{"https://get.threatbite.com/public/free.txt"},
	}

	if configFile == "" {
		if _, err := os.Stat("config_local.env"); !os.IsNotExist(err) {
			configFile = "config_local.env"
		} else if _, err := os.Stat("config.env"); !os.IsNotExist(err) {
			configFile = "config.env"
		}
	}

	if configFile != "" {
		if err := godotenv.Load(configFile); err != nil {
			return nil, err
		}
	}

	if port := os.Getenv("PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("invalid port value: %s, error: %w", port, err)
		}
		config.Port = p
	}

	if debug := os.Getenv("DEBUG"); debug == "true" || debug == "1" {
		config.Debug = true
	}

	if tls := os.Getenv("AUTO_TLS"); tls == "true" || tls == "1" {
		config.AutoTLS = true
	}

	config.PwnedKey = os.Getenv("PWNED_KEY")
	config.MaxmindKey = os.Getenv("MAXMIND_KEY")

	config.SMTPHello = os.Getenv("SMTP_HELLO")
	config.SMTPFrom = os.Getenv("SMTP_FROM")

	lists := map[string]*[]string{
		"PROXY_LIST":          &config.ProxyList,
		"SPAM_LIST":           &config.SpamList,
		"VPN_LIST":            &config.VPNList,
		"DC_LIST":             &config.DCList,
		"EMAIL_DISPOSAL_LIST": &config.EmailDisposalList,
		"EMAIL_FREE_LIST":     &config.EmailFreeList,
	}

	for env, list := range lists {
		if e := os.Getenv(env); e != "" {
			*list = []string{}
			for _, u := range strings.Fields(e) {
				if _, err := url.ParseRequestURI(u); err != nil {
					return nil, fmt.Errorf("invalid list URL: %s, error: %w", u, err)
				}
				*list = append(*list, u)
			}
		}
	}

	return config, nil
}
