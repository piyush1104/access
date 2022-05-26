package client

// Config ...
type Config struct {
	Addr        string `mapstructure:"addr,omitempty"`
	Cert        string `mapstructure:"cert,omitempty"`
	DialTimeout int    `mapstructure:"dial_timeout,omitempty"`
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Addr:        "localhost:8009",
		DialTimeout: DefaultDialTimeout,
	}
}

// SetDefaults set default values for config
func (c *Config) SetDefaults() {
	d := DefaultConfig()
	if c.Addr == "" {
		c.Addr = d.Addr
	}
	if c.DialTimeout == 0 {
		c.DialTimeout = d.DialTimeout
	}
}
