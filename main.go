package main

import (
	"fmt"
	"goload/runner"
)

func main() {
	executor, err := runner.LoadFromYaml("config/spike.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %s\n", err)
	}
	executor.Execute()
}
