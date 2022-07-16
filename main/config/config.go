package config

import (
	"fmt"

	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	Bearer  string `yaml:"bearer"`
	Version string `yaml:"version"`
	Port    string `yaml:"port"`
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

	return yamlConfig

}

func GetNoBearer() interface{} {
	return Bearer{
		Message: "Bearer token not present",
		Status:  "Unauthorized",
		Code:    401,
	}
}

type Bearer struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Code    int    `json:"code"`
}

type Greet struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Code    string `json:"code"`
}
