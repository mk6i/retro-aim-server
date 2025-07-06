package config

//go:generate go run github.com/mk6i/retro-aim-server/cmd/config_generator unix settings.env
type Config struct {
	ApiHost             string `envconfig:"API_HOST" require:"true" val:"127.0.0.1" description:"Specifies the IP address or hostname that the management API binds to for incoming connections (127.0.0.1 restricts to same machine only)."`
	ApiPort             string `envconfig:"API_PORT" required:"true" val:"8080" description:"The port that the management API service binds to."`
	KerberosPort        string `envconfig:"KERBEROS_PORT" required:"true" val:"1088" description:"The port that the Kerberos server binds to."`
	DBPath              string `envconfig:"DB_PATH" required:"true" val:"oscar.sqlite" description:"The path to the SQLite database file. The file and DB schema are auto-created if they doesn't exist."`
	DisableAuth         bool   `envconfig:"DISABLE_AUTH" required:"true" val:"true" description:"Disable password check and auto-create new users at login time. Useful for quickly creating new accounts during development without having to register new users via the management API."`
	LogLevel            string `envconfig:"LOG_LEVEL" required:"true" val:"info" description:"Set logging granularity. Possible values: 'trace', 'debug', 'info', 'warn', 'error'."`
	TOCHost             string `envconfig:"TOC_HOST" require:"true" val:"0.0.0.0" description:"Specifies the IP address or hostname that the TOC service binds to for incoming connections (0.0.0.0 listens on all interfaces)."`
	TOCPort             string `envconfig:"TOC_PORT" required:"true" val:"9898" description:"The port that the TOC service binds to."`
	Listeners           string `envconfig:"LISTENERS" required:"true" val:"EXTERNAL://0.0.0.0:5190" description:"tbd"`
	AdvertisedListeners string `envconfig:"ADVERTISED_LISTENERS" required:"true" val:"EXTERNAL://127.0.0.1:5190" description:"tbd"`
}

type Build struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}
