package factory

type option func(*dbConfig) error

func WithMongoDB(uri, db string) option {
	return func(c *dbConfig) error {
		if c.mongo == nil {
			c.mongo = &mongodbConfig{}
		}
		c.mongo.DB = db
		c.mongo.URI = uri
		return nil
	}
}
