package config

import (
	"context"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	"os"
)

func Get(ctx context.Context, path string) (*Config, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("Path", path)
	log.Info("reading config file")
	f, err := os.Open(path)
	if err != nil {
		log.Error(err, "failed to read file")
	}
	var c Config
	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		log.Error(err, "failed to decode yaml")
		return nil, err
	}
	return &c, nil
}
