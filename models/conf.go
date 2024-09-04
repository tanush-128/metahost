package models


type Config struct {
	Name        string `yaml:"name"`
	Domain      string `yaml:"domain"`
	Description string `yaml:"description"`
	Image       string `yaml:"image"`
	Port        int    `yaml:"port"`
	Nginx       Nginx  `yaml:"nginx"`
}

type Nginx struct {
	Enabled bool   `yaml:"enabled"`
	Conf    string `yaml:"conf"`
}


