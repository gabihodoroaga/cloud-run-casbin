package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

var config *Config

// LoadConfig reads the configuration variables
func LoadConfig() error {
	config = &Config{}

	config.ConnString = os.ExpandEnv(os.Getenv("POSTGRESQL_URL"))
	config.ClientID = os.Getenv("GOOGLE_CLIENT_ID")

	// update the API_PATH from environment variable
	// alternatively this cen be done from docker statup script
	if err := replaceInFile("public/main.js","API_BASE_PATH",fmt.Sprintf(`"%s"`,os.Getenv("API_BASE_PATH"))); err != nil {
		return err
	}
	return nil
}

// Config contains all the application configuration
type Config struct {
	ConnString string
	ClientID   string
}

// GetConfig returns the current application configuration settings
func GetConfig() *Config {
	return config
}

func replaceInFile(file, find, replace string) error {
	input, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	output := bytes.Replace(input, []byte(find), []byte(replace), -1)

	if err = ioutil.WriteFile(file, output, 0666); err != nil {
		return err
	}

	return nil
}
