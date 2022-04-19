package config

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

func Get(path string) (*Config, error) {
	log.Infof("reading config file: %s", path)
	f, err := os.Open(path)
	if err != nil {
		log.WithError(err).Error("failed to read file")
	}
	var c Config
	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		log.WithError(err).Error("failed to decode yaml")
		return nil, err
	}
	return &c, nil
}
