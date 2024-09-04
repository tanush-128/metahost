package tests

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/tanush-128/metahost/models"
	"gopkg.in/yaml.v3"
)

func main() {
	yamlFile, err := ioutil.ReadFile("test.yaml")
	if err != nil {
		log.Fatalf("error reading YAML file: %v", err)
	}

	// Unmarshal the YAML into the struct
	var config models.Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("error unmarshalling YAML: %v", err)
	}

	fmt.Printf("Name: %s\n", config.Name)
	fmt.Printf("Domain: %s\n", config.Domain)
	fmt.Printf("Description: %s\n", config.Description)
	fmt.Printf("Image: %s\n", config.Image)
	fmt.Printf("Port: %d\n", config.Port)
	fmt.Printf("Nginx Enabled: %t\n", config.Nginx.Enabled)
	fmt.Printf("Nginx Conf: \n%s\n", config.Nginx.Conf)
}
