package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type SqsCommands struct {
	AttributeValue string   `yaml:"attributeValue"`
	Command        string   `yaml:"command"`
	Args           []string `yaml:"args"`
}

type Sqs struct {
	QueueName           string        `yaml:"queueName"`
	AttributeName       string        `yaml:"attributeName"`
	Commands            []SqsCommands `yaml:"commands"`
	WaitTimeSeconds     int64         `yaml:"waitTimeSeconds"`
	MaxNumberOfMessages int64         `yaml:"maxNumberOfMessages"`
	WaitBetweenRequest  int64         `yaml:"waitBetweenRequest"`
}

type InstanceID struct {
	AWS bool `yaml:"aws"`
}

type Config struct {
	Sqs        []Sqs      `yaml:"sqs"`
	InstanceID InstanceID `yaml:"instanceID"`
}

func NewConfigFromArgs(args []string) (*Config, error) {
	var file string
	if len(args) >= 1 {
		file = args[0]
	}
	if file == "" || !FileExists(file) {
		return nil, fmt.Errorf("config file not found: %s", file)
	}
	return NewConfig(file)
}

func NewConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	config := &Config{}
	err = yaml.NewDecoder(f).Decode(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
