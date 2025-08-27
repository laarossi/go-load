package main

import (
	"fmt"
	"goload/internal/runner"
)

func main() {
	executor, err := runner.LoadFromYaml("config/spike.yml")
	if err != nil {
		fmt.Printf("Failed to load config: %s\n", err)
	}
	executor.Execute()
}
