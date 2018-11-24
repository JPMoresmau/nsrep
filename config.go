package main

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"

	item "github.com/JPMoresmau/metarep/item"
)

// Config holds the configuration
type Config struct {
	Cassandra item.Cassandra
}

// ReadFileConfig reads configuration from file
func ReadFileConfig(path string) (Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("ReadFileConfig: %v", err)
		return Config{}, err
	}
	return ReadConfig(data)
}

// ReadConfig reads configuration from data
func ReadConfig(data []byte) (Config, error) {
	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("ReadConfig: %v", err)
	}
	return config, err
}
