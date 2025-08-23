package main

import (
	"fmt"
	goload "goload/runner"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func main() {
	configPath := filepath.Join("config", "smoke.yml")
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return
	}

	var collection goload.Collection
	err = yaml.Unmarshal(yamlFile, &collection)
	if err != nil {
		fmt.Printf("Error parsing YAML: %s\n", err)
		return
	}

	executor := goload.Executor{Collection: collection}
	executor.Execute()
}
