package config

import "fmt"

//go:generate go run github.com/mk6i/retro-aim-server/cmd/config_generator windows settings.bat
//go:generate go run github.com/mk6i/retro-aim-server/cmd/config_generator unix settings.env
type Config struct {
	BOSPort     int    `envconfig:"BOS_PORT" default:"5191" description:"The port that the BOS service binds to."`
	BUCPPort    int    `envconfig:"BUCP_PORT" default:"5190" description:"The port that the auth service binds to."`
	ChatPort    int    `envconfig:"CHAT_PORT" default:"5192" description:"The port that the chat service binds to."`
	DBPath      string `envconfig:"DB_PATH" default:"oscar.sqlite" description:"The path to the SQLite database file. The file and DB schema are auto-created if they doesn't exist."`
	DisableAuth bool   `envconfig:"DISABLE_AUTH" default:"true" description:"Disable password check and auto-create new users at login time. Useful for quickly creating new accounts during development without having to register new users via the management API."`
	FailFast    bool   `envconfig:"FAIL_FAST" default:"false" description:"Crash the server in case it encounters a client message type it doesn't recognize. This makes failures obvious for debugging purposes."`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info" description:"Set logging granularity. Possible values: 'trace', 'debug', 'info', 'warn', 'error'."`
	OSCARHost   string `envconfig:"OSCAR_HOST" default:"127.0.0.1" description:"The hostname that AIM clients connect to in order to reach OSCAR services (BOS, BUCP, chat, etc). Make sure the hostname is reachable by all clients. For local development, the default loopback address should work provided the server and AIM client(s) are running on the same machine. For LAN-only clients, a private IP address (e.g. 192.168..) or hostname should suffice. For clients connecting over the Internet, specify your public IP address and ensure that TCP ports 5190-5192 are open on your firewall."`
}

func Address(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
