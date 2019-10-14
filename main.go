package main

import (
	"github.com/jarlex/gommander/command"
)

// This is an example main

func main() {
	config := gommander.Read("plan")
	config.Plan.Execute()
}
