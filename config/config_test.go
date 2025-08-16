package config

import (
	"testing"
)

func TestParseListenersCfg(t *testing.T) {
	tests := []struct {
		name                   string
		bosListeners           string
		bosAdvertisedListeners string
		kerberosListeners      string
		want                   []Listener
		wantErr                bool
		errContains            string
	}{
		{
			name:                   "valid single listener with kerberos",
			bosListeners:           "LOCAL://0.0.0.0:5190",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "LOCAL://0.0.0.0:1088",
			want: []Listener{
				{
					BOSListenAddress:      "0.0.0.0:5190",
					BOSAdvertisedHost:     "127.0.0.1:5190",
					KerberosListenAddress: "0.0.0.0:1088",
				},
			},
			wantErr: false,
		},
		{
			name:                   "valid single listener without kerberos",
			bosListeners:           "LOCAL://0.0.0.0:5190",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "",
			want: []Listener{
				{
					BOSListenAddress:      "0.0.0.0:5190",
					BOSAdvertisedHost:     "127.0.0.1:5190",
					KerberosListenAddress: "",
				},
			},
			wantErr: false,
		},
		{
			name:                   "valid multiple listeners with mixed kerberos",
			bosListeners:           "LAN://192.168.1.10:5190,WAN://0.0.0.0:5191",
			bosAdvertisedListeners: "LAN://192.168.1.10:5190,WAN://example.com:5191",
			kerberosListeners:      "LAN://192.168.1.10:1088",
			want: []Listener{
				{
					BOSListenAddress:      "192.168.1.10:5190",
					BOSAdvertisedHost:     "192.168.1.10:5190",
					KerberosListenAddress: "192.168.1.10:1088",
				},
				{
					BOSListenAddress:      "0.0.0.0:5191",
					BOSAdvertisedHost:     "example.com:5191",
					KerberosListenAddress: "",
				},
			},
			wantErr: false,
		},
		{
			name:                   "missing BOS advertised host",
			bosListeners:           "LOCAL://0.0.0.0:5190",
			bosAdvertisedListeners: "",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "missing BOS advertise address",
		},
		{
			name:                   "missing BOS listen address",
			bosListeners:           "",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "missing BOS listen address for listener `local://`",
		},
		{
			name:                   "duplicate listener definition in BOS",
			bosListeners:           "LOCAL://0.0.0.0:5190,LOCAL://0.0.0.0:5191",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "duplicate listener definition",
		},
		{
			name:                   "duplicate listener definition in advertised",
			bosListeners:           "LOCAL://0.0.0.0:5190",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190,LOCAL://127.0.0.1:5191",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "duplicate listener definition",
		},
		{
			name:                   "duplicate listener definition in kerberos",
			bosListeners:           "LOCAL://0.0.0.0:5190",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "LOCAL://0.0.0.0:1088,LOCAL://0.0.0.0:1089",
			want:                   nil,
			wantErr:                true,
			errContains:            "duplicate listener definition",
		},
		{
			name:                   "invalid URI format in BOS",
			bosListeners:           "invalid-uri",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "missing scheme. Valid format",
		},
		{
			name:                   "invalid URI format in advertised",
			bosListeners:           "LOCAL://0.0.0.0:5190",
			bosAdvertisedListeners: "invalid-uri",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "missing scheme. Valid format",
		},
		{
			name:                   "invalid URI format in kerberos",
			bosListeners:           "LOCAL://0.0.0.0:5190",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "invalid-uri",
			want:                   nil,
			wantErr:                true,
			errContains:            "missing scheme. Valid format",
		},
		{
			name:                   "URI with underscore in scheme",
			bosListeners:           "LOCAL_://0.0.0.0:5190",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "Valid format: SCHEME://HOST:PORT",
		},
		{
			name:                   "BOS listener missing port",
			bosListeners:           "LOCAL://0.0.0.0",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "missing port",
		},
		{
			name:                   "BOS listener missing host",
			bosListeners:           "LOCAL://:5190",
			bosAdvertisedListeners: "LOCAL://127.0.0.1:5190",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "missing host",
		},
		{
			name:                   "complex multi-listener setup",
			bosListeners:           "LAN://192.168.1.10:5190,WAN://0.0.0.0:5191,DOCKER://172.17.0.1:5192",
			bosAdvertisedListeners: "DOCKER://172.17.0.1:5192,LAN://192.168.1.10:5190,WAN://example.com:5191",
			kerberosListeners:      "WAN://0.0.0.0:1089,LAN://192.168.1.10:1088",
			want: []Listener{
				{
					BOSListenAddress:      "192.168.1.10:5190",
					BOSAdvertisedHost:     "192.168.1.10:5190",
					KerberosListenAddress: "192.168.1.10:1088",
				},
				{
					BOSListenAddress:      "0.0.0.0:5191",
					BOSAdvertisedHost:     "example.com:5191",
					KerberosListenAddress: "0.0.0.0:1089",
				},
				{
					BOSListenAddress:      "172.17.0.1:5192",
					BOSAdvertisedHost:     "172.17.0.1:5192",
					KerberosListenAddress: "",
				},
			},
			wantErr: false,
		},
		{
			name:                   "empty strings for all inputs",
			bosListeners:           "",
			bosAdvertisedListeners: "",
			kerberosListeners:      "",
			want:                   nil,
			wantErr:                true,
			errContains:            "at least one BOS listener is required",
		},
		{
			name:                   "whitespace-only strings",
			bosListeners:           "   ",
			bosAdvertisedListeners: "   ",
			kerberosListeners:      "   ",
			want:                   nil,
			wantErr:                true,
			errContains:            "at least one BOS listener is required",
		},
		{
			name:                   "only kerberos listeners provided",
			bosListeners:           "",
			bosAdvertisedListeners: "",
			kerberosListeners:      "LOCAL://0.0.0.0:1088",
			want:                   nil,
			wantErr:                true,
			errContains:            "missing BOS advertise address for listener `local://`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				BOSListeners:       tt.bosListeners,
				BOSAdvertisedHosts: tt.bosAdvertisedListeners,
				KerberosListeners:  tt.kerberosListeners,
			}
			got, err := config.ParseListenersCfg()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseListenersCfg() expected error but got none")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseListenersCfg() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseListenersCfg() unexpected error = %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParseListenersCfg() returned %d listeners, want %d", len(got), len(tt.want))
				return
			}

			// Create maps for easier comparison
			gotMap := make(map[string]Listener)
			wantMap := make(map[string]Listener)

			for _, l := range got {
				key := l.BOSListenAddress + "|" + l.BOSAdvertisedHost
				gotMap[key] = l
			}

			for _, l := range tt.want {
				key := l.BOSListenAddress + "|" + l.BOSAdvertisedHost
				wantMap[key] = l
			}

			for key, wantListener := range wantMap {
				gotListener, exists := gotMap[key]
				if !exists {
					t.Errorf("ParseListenersCfg() missing listener with key %s", key)
					continue
				}

				if gotListener.BOSListenAddress != wantListener.BOSListenAddress {
					t.Errorf("ParseListenersCfg() BOSListenAddress = %v, want %v", gotListener.BOSListenAddress, wantListener.BOSListenAddress)
				}
				if gotListener.BOSAdvertisedHost != wantListener.BOSAdvertisedHost {
					t.Errorf("ParseListenersCfg() BOSAdvertisedHost = %v, want %v", gotListener.BOSAdvertisedHost, wantListener.BOSAdvertisedHost)
				}
				if gotListener.KerberosListenAddress != wantListener.KerberosListenAddress {
					t.Errorf("ParseListenersCfg() KerberosListenAddress = %v, want %v", gotListener.KerberosListenAddress, wantListener.KerberosListenAddress)
				}
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config with all fields",
			config: Config{
				TOCListeners: "0.0.0.0:9898,192.168.1.10:9899",
				APIListener:  "127.0.0.1:8080",
			},
			wantErr: false,
		},
		{
			name: "valid config with single TOC listener",
			config: Config{
				TOCListeners: "0.0.0.0:9898",
				APIListener:  "127.0.0.1:8080",
			},
			wantErr: false,
		},
		{
			name: "valid config with empty TOC listeners",
			config: Config{
				TOCListeners: "",
				APIListener:  "127.0.0.1:8080",
			},
			wantErr: false,
		},
		{
			name: "valid config with empty API listener",
			config: Config{
				TOCListeners: "0.0.0.0:9898",
				APIListener:  "",
			},
			wantErr:     true,
			errContains: "APIListener is required and cannot be empty",
		},
		{
			name: "valid config with all empty",
			config: Config{
				TOCListeners: "",
				APIListener:  "",
			},
			wantErr:     true,
			errContains: "APIListener is required and cannot be empty",
		},
		{
			name: "invalid TOC listener - missing port",
			config: Config{
				TOCListeners: "0.0.0.0",
				APIListener:  "127.0.0.1:8080",
			},
			wantErr:     true,
			errContains: "invalid TOC listener \"0.0.0.0\": address 0.0.0.0: missing port in address",
		},
		{
			name: "invalid TOC listener - missing host",
			config: Config{
				TOCListeners: ":9898",
				APIListener:  "127.0.0.1:8080",
			},
			wantErr:     true,
			errContains: "invalid TOC listener \":9898\": missing host",
		},
		{
			name: "invalid TOC listener - malformed",
			config: Config{
				TOCListeners: "invalid-format",
				APIListener:  "127.0.0.1:8080",
			},
			wantErr:     true,
			errContains: "invalid TOC listener \"invalid-format\": address invalid-format: missing port in address",
		},
		{
			name: "invalid TOC listener in comma-separated list",
			config: Config{
				TOCListeners: "0.0.0.0:9898,invalid-format,192.168.1.10:9899",
				APIListener:  "127.0.0.1:8080",
			},
			wantErr:     true,
			errContains: "invalid TOC listener \"invalid-format\": address invalid-format: missing port in address",
		},
		{
			name: "invalid API listener - missing port",
			config: Config{
				TOCListeners: "0.0.0.0:9898",
				APIListener:  "127.0.0.1",
			},
			wantErr:     true,
			errContains: "invalid API listener \"127.0.0.1\": address 127.0.0.1: missing port in address",
		},
		{
			name: "invalid API listener - missing host",
			config: Config{
				TOCListeners: "0.0.0.0:9898",
				APIListener:  ":8080",
			},
			wantErr:     true,
			errContains: "invalid API listener \":8080\": missing host",
		},
		{
			name: "invalid API listener - malformed",
			config: Config{
				TOCListeners: "0.0.0.0:9898",
				APIListener:  "invalid-format",
			},
			wantErr:     true,
			errContains: "invalid API listener \"invalid-format\": address invalid-format: missing port in address",
		},
		{
			name: "whitespace-only TOC listeners",
			config: Config{
				TOCListeners: "   ,  ,  ",
				APIListener:  "127.0.0.1:8080",
			},
			wantErr: false,
		},
		{
			name: "whitespace-only API listener",
			config: Config{
				TOCListeners: "0.0.0.0:9898",
				APIListener:  "   ",
			},
			wantErr:     true,
			errContains: "APIListener is required and cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got none")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Config.Validate() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Config.Validate() unexpected error = %v", err)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
