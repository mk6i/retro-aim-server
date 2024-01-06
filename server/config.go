package server

import "fmt"

type Config struct {
	BOSPort     int    `envconfig:"BOS_PORT" default:"5191"`
	BUCPPort    int    `envconfig:"BUCP_PORT" default:"5190"`
	ChatPort    int    `envconfig:"CHAT_PORT" default:"5192"`
	DBPath      string `envconfig:"DB_PATH" required:"true"`
	DisableAuth bool   `envconfig:"DISABLE_AUTH" default:"false"`
	FailFast    bool   `envconfig:"FAIL_FAST" default:"false"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	OSCARHost   string `envconfig:"OSCAR_HOST" required:"true"`
}

func Address(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
