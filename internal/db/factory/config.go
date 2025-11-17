package factory

type mongodbConfig struct {
	URI string
	DB  string
}

type dbConfig struct {
	mongo *mongodbConfig
}

var defaultConfig = &dbConfig{
	mongo: &mongodbConfig{
		URI: "mongodb://localhost:27017",
		DB:  "payment",
	},
}
