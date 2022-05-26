package server

// Config for rpcpolicy node

func (config *Config) SetDefaults() {
	d := DefaultConfig()

	if config.Port == 0 {
		config.Port = d.Port
	}
	if config.Metrics == 0 {
		config.Metrics = d.Metrics
	} else if config.Metrics == -1 {
		config.Metrics = 0
	}
}
