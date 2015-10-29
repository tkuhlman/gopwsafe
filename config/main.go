package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Config object defining the configuration for gopwsafe
type Config struct {
	History       []string `yaml:",omitempty"`
	HistoryLength int
}

// PWSafeDBConfig An interface that defines various methods for interacting with the pwsafe configuration
type PWSafeDBConfig interface {
	AddToPathHistory(string) error
	GetPathHistory() []string
	Save() error
}

// setDefaults sets configuration defaults
func (conf *Config) setDefaults() {
	conf.HistoryLength = 5
}

// Load the config from the standard location
func Load() PWSafeDBConfig {
	var conf Config
	conf.setDefaults()
	//todo Allow configuring a config file
	configPath := os.Getenv("HOME") + "/.gopwsafe.yaml"

	// Before trying to load the file if it doesn't exist return
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &conf
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	return &conf
}

// Save the configuration to disk
func (conf *Config) Save() error {
	//todo Allow configuring a config file
	configPath := os.Getenv("HOME") + "/.gopwsafe.yaml"
	data, err := yaml.Marshal(&conf)
	if err != nil {
		return err
	}
	ioutil.WriteFile(configPath, data, 0640)
	if err != nil {
		return err
	}

	return nil
}
