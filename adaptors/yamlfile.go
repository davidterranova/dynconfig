package adaptors

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type YAMLFileAdatptor struct {
	filePath string
}

func NewYAMLFileAdatptor(path string) YAMLFileAdatptor {
	return YAMLFileAdatptor{filePath: path}
}

func (a YAMLFileAdatptor) Read(config interface{}) error {
	f, err := os.Open(a.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("cannot read file: %w", err)
		}
		return nil
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	return decoder.Decode(config)
}

func (a YAMLFileAdatptor) Write(config interface{}) error {
	f, err := os.OpenFile(a.filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}
	defer f.Close()
	encoder := yaml.NewEncoder(f)
	return encoder.Encode(config)
}
