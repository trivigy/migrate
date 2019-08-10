package types

import (
	"fmt"
	"net/url"
	"strings"
)

var (
	validSSLModes = []string{"disable", "require", "verify-ca", "verify-full"}
	escaper       = strings.NewReplacer(` `, `\ `, `'`, `\'`, `\`, `\\`)
)

// PsqlDSN config represents a parsed PostgreSQL connection string.
type PsqlDSN struct {
	Host                    string `json:"host" yaml:"host"`
	Port                    string `json:"port" yaml:"port"`
	User                    string `json:"user" yaml:"user"`
	Password                string `json:"password" yaml:"password"`
	DBName                  string `json:"dbName" yaml:"dbName"`
	SSLMode                 string `json:"sslMode" yaml:"sslMode"`
	FallbackApplicationName string `json:"fallbackApplicationName" yaml:"fallbackApplicationName"`
	ConnectTimeout          int    `json:"connectTimeout" yaml:"connectTimeout"`
	SSLCert                 string `json:"sslCert" yaml:"sslCert"`
	SSLKey                  string `json:"sslKey" yaml:"sslKey"`
	SSLRootCert             string `json:"sslRootCert" yaml:"sslRootCert"`
}

// Driver returns the driver name for PostgresSQL connection string.
func (c PsqlDSN) Driver() string {
	return "postgres"
}

// Source reassembles the parsed PostgreSQL connection string.
func (c PsqlDSN) Source() string {
	u := url.URL{Scheme: "postgres"}

	// u := make([]string, 0)

	var host string
	if c.Host != "" {
		host = escaper.Replace(c.Host)
	}
	if c.Port != "" {
		host += ":" + escaper.Replace(c.Port)
	}
	u.Host = host

	if c.Password != "" {
		if c.User != "" {
			u.User = url.UserPassword(escaper.Replace(c.User), escaper.Replace(c.Password))
		}
	} else {
		if c.User != "" {
			u.User = url.User(escaper.Replace(c.User))
		}
	}

	if c.DBName != "" {
		u.Path = "/" + escaper.Replace(c.DBName)
	}

	q := u.Query()
	if c.SSLMode != "" {
		for _, valid := range validSSLModes {
			if valid == c.SSLMode {
				q.Set("sslmode", escaper.Replace(c.SSLMode))
			}
		}
	} else {
		q.Set("sslmode", "disable")
	}

	if c.FallbackApplicationName != "" {
		q.Set("fallback_application_name", escaper.Replace(c.FallbackApplicationName))
	}

	if c.ConnectTimeout != 0 {
		q.Set("connect_timeout", fmt.Sprintf("%d", c.ConnectTimeout))
	}

	if c.SSLCert != "" {
		q.Set("sslcert", escaper.Replace(c.SSLCert))
	}

	if c.SSLKey != "" {
		q.Set("sslkey", escaper.Replace(c.SSLKey))
	}

	if c.SSLRootCert != "" {
		q.Set("sslrootcert", escaper.Replace(c.SSLRootCert))
	}
	u.RawQuery = q.Encode()
	return u.String()
}
