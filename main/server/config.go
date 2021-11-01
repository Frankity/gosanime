package server

import (
	"fmt"

	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	Bearer string `yaml:"bearer"`
}

func Config() YamlConfig {

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
	}

	var yamlConfig YamlConfig
	err = yaml.Unmarshal(yamlFile, &yamlConfig)
	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", err)
	}

	fmt.Printf("Result: %v\n", yamlConfig.Bearer)

	return yamlConfig

}
