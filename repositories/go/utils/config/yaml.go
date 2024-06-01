package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

func ReadYamlFile(file string, body any) error {
	bs, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bs, body)
	return err
}
