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
	BOSListeners       string `envconfig:"OSCAR_LISTENERS" required:"true" val:"LOCAL://0.0.0.0:5190" description:"Network listeners for core OSCAR services. For multi-homed servers, allows users to connect from multiple networks. For example, you can allow both LAN and Internet clients to connect to the same server using different connection settings.\n\nFormat:\n\t- Comma-separated list of [NAME]://[HOSTNAME]:[PORT]\n\t- Listener names and ports must be unique\n\t- Listener names are user-defined\n\t- Each listener needs OSCAR_ADVERTISED_LISTENERS/KERBEROS_LISTENERS configs\n\nExamples:\n\t// Listen on all interfaces\n\tLAN://0.0.0.0:5190\n\t// Separate Internet and LAN config\n\tWAN://142.250.176.206:5190,LAN://192.168.1.10:5191"`
	BOSAdvertisedHosts string `envconfig:"OSCAR_ADVERTISED_LISTENERS" required:"true" val:"LOCAL://127.0.0.1:5190" description:"Hostnames published by the server that clients connect to for accessing various OSCAR services. These hostnames are NOT the bind addresses. For multi-homed use servers, allows clients to connect using separate hostnames per network.\n\nFormat:\n\t- Comma-separated list of [NAME]://[HOSTNAME]:[PORT]\n\t- Each listener config must correspond to a config in OSCAR_LISTENERS\n\t- Clients MUST be able to connect to these hostnames\n\nExamples:\n\t// Local LAN config, server behind NAT\n\tLAN://0.0.0.0:5190\n\t// Separate Internet and LAN config\n\tWAN://aim.example.com:5190,LAN://192.168.1.10:5191"`
	KerberosListeners  string `envconfig:"KERBEROS_LISTENERS" required:"true" val:"LOCAL://0.0.0.0:1088" description:"Network listeners for Kerberos authentication. See OSCAR_LISTENERS doc for more details.\n\nExamples:\n\t// Listen on all interfaces\n\tLAN://0.0.0.0:1088\n\t// Separate Internet and LAN config\n\tWAN://142.250.176.206:1088,LAN://192.168.1.10:1087"`
	TOCListeners       string `envconfig:"TOC_LISTENERS" required:"true" val:"0.0.0.0:9898" description:"Network listeners for TOC protocol service.\n\nFormat: Comma-separated list of hostname:port pairs.\n\nExamples:\n\t// All interfaces\n\t0.0.0.0:9898\n\t// Multiple listeners\n\t0.0.0.0:9898,192.168.1.10:9899"`
	APIListener        string `envconfig:"API_LISTENER" required:"true" val:"127.0.0.1:8080" description:"Network listener for management API binds to. Only 1 listener can be specified. (Default 127.0.0.1 restricts to same machine only)."`

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
	BOSAdvertisedHost     string
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
		if m[u.Scheme].BOSAdvertisedHost != "" {
			return nil, errors.New("duplicate listener definition")
		}
		m[u.Scheme].BOSAdvertisedHost = net.JoinHostPort(u.Hostname(), u.Port())
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
		case v.BOSAdvertisedHost == "":
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
