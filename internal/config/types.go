package config

type Config struct {
	Images []Image `yaml:"images"`
}

type Image struct {
	Source      string
	Destination string
}
