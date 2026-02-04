package factory

import "go.opentelemetry.io/otel/trace"

type mongodbConfig struct {
	URI         string
	DB          string
	MaxPoolSize uint64
	MinPoolSize uint64
}

type dbConfig struct {
	mongo  *mongodbConfig
	tracer trace.Tracer
}

var defaultConfig = &dbConfig{
	mongo: &mongodbConfig{
		URI:         "mongodb://localhost:27017",
		DB:          "payment",
		MaxPoolSize: 30,
		MinPoolSize: 10,
	},
}
