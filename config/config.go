package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

//go:generate go run github.com/mk6i/retro-aim-server/cmd/config_generator unix settings.env
type Config struct {
	APIListener        string `envconfig:"API_LISTENERS" required:"true" val:"0.0.0.0:8080" description:"tbd"`
	BOSAdvertisedHosts string `envconfig:"BOS_ADVERTISED_HOSTS" required:"true" val:"EXTERNAL://127.0.0.1:5190" description:"tbd"`
	BOSListeners       string `envconfig:"BOS_LISTENERS" required:"true" val:"EXTERNAL://0.0.0.0:5190" description:"tbd"`
	KerberosListeners  string `envconfig:"KERBEROS_LISTENERS" required:"true" val:"EXTERNAL://0.0.0.0:1088" description:"tbd"`
	TOCListeners       string `envconfig:"TOC_LISTENERS" required:"true" val:"0.0.0.0:9898" description:"tbd"`

	DBPath      string `envconfig:"DB_PATH" required:"true" val:"oscar.sqlite" description:"The path to the SQLite database file. The file and DB schema are auto-created if they doesn't exist."`
	DisableAuth bool   `envconfig:"DISABLE_AUTH" required:"true" val:"true" description:"Disable password check and auto-create new users at login time. Useful for quickly creating new accounts during development without having to register new users via the management API."`
	LogLevel    string `envconfig:"LOG_LEVEL" required:"true" val:"info" description:"Set logging granularity. Possible values: 'trace', 'debug', 'info', 'warn', 'error'."`
}

type Build struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

type Listener struct {
	BOSListenAddress      string
	BOSAdvertisedHosts    string
	KerberosListenAddress string
}

func ParseListenersCfg(BOSListeners string, BOSAdvertisedListeners string, kerberosListeners string) ([]Listener, error) {
	m := make(map[string]*Listener)

	for _, lStr := range strings.Split(BOSListeners, ",") {
		u, err := url.Parse(lStr)
		if err != nil {
			return nil, fmt.Errorf("parsing listener URI: %w", err)
		}
		if _, ok := m[u.Scheme]; !ok {
			m[u.Scheme] = &Listener{}
		}
		if m[u.Scheme].BOSListenAddress != "" {
			return nil, errors.New("duplicate listener definition")
		}
		m[u.Scheme].BOSListenAddress = net.JoinHostPort(u.Hostname(), u.Port())
	}

	for _, lStr := range strings.Split(BOSAdvertisedListeners, ",") {
		u, err := url.Parse(lStr)
		if err != nil {
			return nil, fmt.Errorf("parsing listener URI: %w", err)
		}
		if _, ok := m[u.Scheme]; !ok {
			m[u.Scheme] = &Listener{}
		}
		if m[u.Scheme].BOSAdvertisedHosts != "" {
			return nil, errors.New("duplicate listener definition")
		}
		m[u.Scheme].BOSAdvertisedHosts = net.JoinHostPort(u.Hostname(), u.Port())
	}

	for _, lStr := range strings.Split(kerberosListeners, ",") {
		u, err := url.Parse(lStr)
		if err != nil {
			return nil, fmt.Errorf("parsing listener URI: %w", err)
		}
		if _, ok := m[u.Scheme]; !ok {
			m[u.Scheme] = &Listener{}
		}
		if m[u.Scheme].KerberosListenAddress != "" {
			return nil, errors.New("duplicate listener definition")
		}
		m[u.Scheme].KerberosListenAddress = net.JoinHostPort(u.Hostname(), u.Port())
	}

	ret := make([]Listener, 0, len(m))

	for k, v := range m {
		switch {
		case v.BOSAdvertisedHosts == "":
			return nil, fmt.Errorf("missing BOS advertise address for listener `%s://`", k)
		case v.BOSListenAddress == "":
			return nil, fmt.Errorf("missing BOS listen address for listener `%s://`", k)
		case v.KerberosListenAddress == "":
			return nil, fmt.Errorf("missing kerberos listen address for listener `%s://`", k)
		}
		ret = append(ret, *v)
	}

	return ret, nil
}
