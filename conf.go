package main

import (
	"gopkg.in/yaml.v3"
)

type MetaHostConf struct {
	DefaultNginxConf string `yaml:"nginx_conf"`
	DefaultEmail     string `yaml:"email"`
}

func ParseMetaHostConf(data string) (MetaHostConf, error) {
	conf := MetaHostConf{}
	err := yaml.Unmarshal([]byte(data), &conf)
	if err != nil {
		return MetaHostConf{}, err
	}

	return conf, nil
}
