package service

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"
)

type Config struct {
	Address, DatabaseURL                            string
	RequestTimeout, StartupTimeout, ShutdownTimeout time.Duration
}

func LoadConfig(getenv func(string) string) (Config, error) {
	c := Config{Address: getenv("HTTP_ADDR"), DatabaseURL: getenv("DATABASE_URL"), RequestTimeout: 5 * time.Second, StartupTimeout: 10 * time.Second, ShutdownTimeout: 10 * time.Second}
	if c.Address == "" {
		c.Address = ":8103"
	}
	_, port, err := net.SplitHostPort(c.Address)
	if err != nil {
		return c, fmt.Errorf("HTTP_ADDR must be host:port")
	}
	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return c, fmt.Errorf("HTTP_ADDR port must be between 1 and 65535")
	}
	for key, target := range map[string]*time.Duration{"REQUEST_TIMEOUT": &c.RequestTimeout, "STARTUP_TIMEOUT": &c.StartupTimeout, "SHUTDOWN_TIMEOUT": &c.ShutdownTimeout} {
		if raw := getenv(key); raw != "" {
			d, err := time.ParseDuration(raw)
			if err != nil || d <= 0 || d > time.Minute {
				return c, fmt.Errorf("%s must be a positive duration at most 1m", key)
			}
			*target = d
		}
	}
	if c.DatabaseURL != "" {
		u, err := url.Parse(c.DatabaseURL)
		if err != nil || u.Hostname() == "" || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
			return c, fmt.Errorf("DATABASE_URL must be a PostgreSQL URL")
		}
	}
	return c, nil
}
