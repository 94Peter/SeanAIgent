package factory

import "go.opentelemetry.io/otel/trace"

type option func(*dbConfig)

func WithMongoDB(uri, db string) option {
	return func(c *dbConfig) {
		if c.mongo == nil {
			c.mongo = &mongodbConfig{}
		}
		c.mongo.DB = db
		c.mongo.URI = uri

	}
}

func WithTracer(t trace.Tracer) option {
	return func(c *dbConfig) {
		c.tracer = t
	}
}
