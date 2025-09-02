package main

//go:generate sh -c "go run main.go server --print-openapi dummy > openapi.json"

import "github.com/coder/agentapi/cmd"

func main() {
	cmd.Execute()
}
