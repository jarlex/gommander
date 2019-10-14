package main

import (
    "github.com/jarlex/gommander/command"
)

// This is an example main

func main() {
    config := command.Read("plan")
    config.Plan.Execute()
}
