package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey string `yaml:"APIKey"`
	URL    string `yaml:"URL"`
}

func getConfigFile() (*os.File, error) {
	var configFile *os.File
	configDir, err := getConfigDirectory()
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return nil, err
	}
	configFilePath := filepath.Join(configDir, "config.yml")
	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		configFile, err = os.Create(configFilePath)
		if err != nil {
			return nil, err
		}
	} else {
		configFile, err = os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
	}
	return configFile, nil
}

func getConfigDirectory() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "fileshare-client"), nil
}

func getConfig() (*Config, error) {
	file, err := getConfigFile()
	if err != nil {
		log.Printf("Error getting config file")
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading config")
		return nil, err
	}
	config := &Config{}
	err = yaml.Unmarshal(fileData, config)
	if err != nil {
		log.Printf("Error unmarshalling")
		return nil, err
	}
	return config, err
}

func putConfig(config *Config) error {
	file, err := getConfigFile()
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	fileData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = file.Write(fileData)
	if err != nil {
		return err
	}
	return nil
}

func getOrPromptForValue(input *string, name string) error {
	if input == nil || *input == "" {
		return promptForValue(input, name)
	}
	return nil
}

func promptForValue(input *string, name string) error {
	fmt.Printf("Enter %s: ", name)
	_, err := fmt.Scanln(input)
	if err != nil {
		return err
	}
	return nil
}
