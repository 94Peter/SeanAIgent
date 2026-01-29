package factory

import "go.opentelemetry.io/otel/trace"

type mongodbConfig struct {
	URI string
	DB  string
}

type dbConfig struct {
	mongo  *mongodbConfig
	tracer trace.Tracer
}

var defaultConfig = &dbConfig{
	mongo: &mongodbConfig{
		URI: "mongodb://localhost:27017",
		DB:  "payment",
	},
}
