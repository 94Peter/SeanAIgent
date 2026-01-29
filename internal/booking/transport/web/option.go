package web

type Option func(cfg *Config)

func WithPort(port uint16) Option {
	return func(cfg *Config) {
		cfg.port = port
	}
}

func WithSession(enable bool, store string, cookieName string, maxAge int, keyPairs ...string) Option {
	return func(cfg *Config) {
		cfg.session.Enable = enable
		cfg.session.Store = store
		cfg.session.CookieName = cookieName
		cfg.session.MaxAge = maxAge
		cfg.session.KeyPairs = keyPairs
	}
}

func WithCsrf(enable bool, secret string, fieldname string, excludePaths ...string) Option {
	return func(cfg *Config) {
		cfg.csrf.Enable = enable
		cfg.csrf.Secret = secret
		cfg.csrf.FieldName = fieldname
		cfg.csrf.ExcludePaths = excludePaths
	}
}

func WithTracerEnable(enable bool) Option {
	return func(cfg *Config) {
		cfg.tracer.Enable = enable
	}
}

func WithLoggerEnable(enable bool) Option {
	return func(cfg *Config) {
		cfg.logger.Enable = enable
	}
}

func WithMode(mode string) Option {
	return func(cfg *Config) {
		cfg.mode = mode
	}
}
