package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// YamlConfig struct represents the application's configuration loaded from a YAML file.
// It includes settings like API bearer token, version, and port.
type YamlConfig struct {
	Bearer  string `yaml:"bearer"`  // Bearer token required for API authorization.
	Version string `yaml:"version"` // Version of the application/API.
	Port    string `yaml:"port"`    // Port on which the application server will listen.
}

// Config reads the 'config.yaml' file from the current directory, unmarshals it into
// a YamlConfig struct, and returns the struct.
// If reading or parsing the YAML file fails, it prints an error message to stdout
// and returns a partially populated or zero-value YamlConfig struct.
// Consider returning an error from this function for better error handling by the caller.
//
// Returns:
//   YamlConfig: The application configuration populated from 'config.yaml'.
func Config() YamlConfig {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		// Printing directly to stdout is not ideal for library/package code.
		// Consider logging or returning the error.
		fmt.Printf("Error reading YAML file: %s\n", err)
	}

	var yamlConfig YamlConfig
	err = yaml.Unmarshal(yamlFile, &yamlConfig)
	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", err)
	}

	return yamlConfig
}

// GetNoBearer creates and returns a standard error response object for when a Bearer token
// is missing or invalid. This is typically used to inform the client about authorization failure.
//
// Returns:
//   interface{}: A Bearer struct (which will be marshalled to JSON) containing
//                the error message, status, and HTTP status code.
func GetNoBearer() interface{} {
	return Bearer{
		Message: "Bearer token not present",
		Status:  "Unauthorized", // Status text.
		Code:    401,            // HTTP status code for Unauthorized.
	}
}

// Bearer struct is used to create a standardized JSON response for Bearer token related errors.
type Bearer struct {
	Message string `json:"message"` // Error message explaining the issue (e.g., "Bearer token not present").
	Status  string `json:"status"`  // A short status description (e.g., "Unauthorized").
	Code    int    `json:"code"`    // The HTTP status code associated with this error (e.g., 401).
}

// Greet struct is used to create a standardized JSON response for the API's root/index endpoint.
// It provides a welcome message, status, and a code.
type Greet struct {
	Message string `json:"message"` // Welcome message, often including API version (e.g., "Gosanime Api v: 1.0 is Running.").
	Status  string `json:"status"`  // Operational status of the API (e.g., "OK", "Running").
	Code    string `json:"code"`    // A status code, often "200" for successful greeting. Consider using int for consistency.
}
