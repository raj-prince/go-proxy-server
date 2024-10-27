package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type RetryConfig struct {
	Method           string `yaml:"method"`
	FileName         string `yaml:"fileName"`
	RetryInstruction string `yaml:"retryInstruction"`
	RetryCount       int    `yaml:"retryCount"`
	SkipCount        int    `yaml:"skipCount"`
}

type Config struct {
	TargetHost  string        `yaml:"targetHost"`
	RetryConfig []RetryConfig `yaml:"retryConfig"`
}

func parseConfigFile(configPath string) (*Config, error) {
	var config Config

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	// Access the values from the struct
	fmt.Println("Target Host:", config.TargetHost)
	for _, retry := range config.RetryConfig {
		fmt.Println("Method:", retry.Method)
		fmt.Println("File Name:", retry.FileName)
		fmt.Println("Retry instructions:", retry.RetryInstruction)
		fmt.Println("Retry Count:", retry.RetryCount)
		fmt.Println("Skip Count:", retry.SkipCount)
	}

	return &config, nil
}
