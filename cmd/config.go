package cmd

import (
	"seanAIgent/internal/booking/transport/web"

	"github.com/spf13/viper"
)

func ProvideWebConfig() web.Config {
	opts := []web.Option{
		web.WithPort(viper.GetUint16("http.port")),
		web.WithSession(
			viper.GetBool("http.session.enabled"),
			viper.GetString("http.session.store"),
			viper.GetString("http.session.cookie_name"),
			viper.GetInt("http.session.max_age"),
			viper.GetStringSlice("http.session.key_pairs")...,
		),
		web.WithCsrf(
			viper.GetBool("http.csrf.enabled"),
			viper.GetString("http.csrf.secret"),
			viper.GetString("http.csrf.field_name"),
			viper.GetStringSlice("http.csrf.ignore_paths")...,
		),
		web.WithTracerEnable(viper.GetBool("http.tracer.enabled")),
		web.WithLoggerEnable(viper.GetBool("http.logger.enabled")),
		web.WithMode(viper.GetString("http.mode")),
		web.WithMaxConcurrentRequests(viper.GetInt("http.max_concurrent_requests")),
	}

	cfg := web.Config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
