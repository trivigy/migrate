package nub

import (
	"fmt"
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
	u := make([]string, 0)

	if c.Host != "" {
		u = append(u, "host="+escaper.Replace(c.Host))
	}

	if c.Port != "" {
		u = append(u, "port="+escaper.Replace(c.Port))
	} else {
		u = append(u, fmt.Sprintf("port=%d", 5432))
	}

	if c.User != "" {
		u = append(u, "user="+escaper.Replace(c.User))
	}

	if c.Password != "" {
		u = append(u, "password="+escaper.Replace(c.Password))
	}

	if c.DBName != "" {
		u = append(u, "dbname="+escaper.Replace(c.DBName))
	}

	if c.SSLMode != "" {
		for _, valid := range validSSLModes {
			if valid == c.SSLMode {
				u = append(u, "sslmode="+escaper.Replace(c.SSLMode))
			}
		}
	} else {
		u = append(u, "sslmode=disable")
	}

	if c.FallbackApplicationName != "" {
		u = append(u, "fallback_application_name="+escaper.Replace(c.FallbackApplicationName))
	}

	if c.ConnectTimeout != 0 {
		u = append(u, fmt.Sprintf("connect_timeout=%d", c.ConnectTimeout))
	}

	if c.SSLCert != "" {
		u = append(u, "sslcert="+escaper.Replace(c.SSLCert))
	}

	if c.SSLKey != "" {
		u = append(u, "sslkey="+escaper.Replace(c.SSLKey))
	}

	if c.SSLRootCert != "" {
		u = append(u, "sslrootcert="+escaper.Replace(c.SSLRootCert))
	}

	if len(u) == 0 {
		return ""
	}

	return strings.Join(u, " ")
}
