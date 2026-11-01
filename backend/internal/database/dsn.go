package database

import (
	"net"
	"net/url"
	"strconv"

	"github.com/black-coda/go-boilerplate/internal/config"
)

func DsnConfig(config *config.Config) *url.URL {
	return &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(config.Database.User, config.Database.Password),
		Host:     net.JoinHostPort(config.Database.Host, strconv.Itoa(config.Database.Port)),
		Path:     config.Database.Name,
		RawQuery: "sslmode=" + url.QueryEscape(config.Database.SSLMode),
	}
}
