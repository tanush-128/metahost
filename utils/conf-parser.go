package utils

import (
	"github.com/tanush-128/metahost/models"
	"gopkg.in/yaml.v3"
)

func ParseYaml(data string) (models.Config, error) {
	conf := models.Config{}
	err := yaml.Unmarshal([]byte(data), &conf)
	if err != nil {
		return models.Config{}, err
	}

	return conf, nil
}
