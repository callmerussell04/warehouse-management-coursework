package app

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type config struct {
	jwtSecret      []byte
	adminUsername  string
	adminEmail     string
	adminPassword  string
	allowedOrigins []string
	cookieSecure   bool
	trustedProxies []string
}

func loadConfig() (config, error) {
	required := func(name string) (string, error) {
		value := os.Getenv(name)
		if strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("%s is required", name)
		}
		return value, nil
	}

	jwtSecret, err := required("JWT_SECRET")
	if err != nil {
		return config{}, err
	}
	if len(jwtSecret) < 32 {
		return config{}, fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}

	adminUsername, err := required("ADMIN_USERNAME")
	if err != nil {
		return config{}, err
	}
	adminUsername = strings.TrimSpace(adminUsername)
	if len([]rune(adminUsername)) > 255 {
		return config{}, fmt.Errorf("ADMIN_USERNAME must contain at most 255 characters")
	}
	adminEmail, err := required("ADMIN_EMAIL")
	if err != nil {
		return config{}, err
	}
	adminEmail = strings.ToLower(strings.TrimSpace(adminEmail))
	parsedEmail, parseEmailErr := mail.ParseAddress(adminEmail)
	if len(adminEmail) > 254 || parseEmailErr != nil || parsedEmail.Address != adminEmail {
		return config{}, fmt.Errorf("ADMIN_EMAIL must be a valid email address of at most 254 bytes")
	}
	adminPassword, err := required("ADMIN_PASSWORD")
	if err != nil {
		return config{}, err
	}
	if len(adminPassword) < 12 || len(adminPassword) > 72 {
		return config{}, fmt.Errorf("ADMIN_PASSWORD must contain between 12 and 72 bytes")
	}

	originsValue, err := required("ALLOWED_ORIGINS")
	if err != nil {
		return config{}, err
	}
	origins, err := parseOrigins(originsValue)
	if err != nil {
		return config{}, err
	}

	cookieSecureValue, err := required("COOKIE_SECURE")
	if err != nil {
		return config{}, err
	}
	cookieSecure, err := strconv.ParseBool(cookieSecureValue)
	if err != nil {
		return config{}, fmt.Errorf("COOKIE_SECURE must be true or false")
	}

	trustedProxies, err := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	if err != nil {
		return config{}, err
	}

	return config{
		jwtSecret:      []byte(jwtSecret),
		adminUsername:  adminUsername,
		adminEmail:     adminEmail,
		adminPassword:  adminPassword,
		allowedOrigins: origins,
		cookieSecure:   cookieSecure,
		trustedProxies: trustedProxies,
	}, nil
}

func parseTrustedProxies(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	proxies := make([]string, 0)
	for _, raw := range strings.Split(value, ",") {
		proxy := strings.TrimSpace(raw)
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				return nil, fmt.Errorf("TRUSTED_PROXIES contains invalid IP address or CIDR %q", proxy)
			}
		}
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

func parseOrigins(value string) ([]string, error) {
	seen := make(map[string]struct{})
	origins := make([]string, 0)
	for _, raw := range strings.Split(value, ",") {
		origin := strings.TrimSpace(raw)
		u, err := url.Parse(origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("ALLOWED_ORIGINS contains invalid origin %q", origin)
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	if len(origins) == 0 {
		return nil, fmt.Errorf("ALLOWED_ORIGINS must contain at least one origin")
	}
	return origins, nil
}
