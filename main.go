package main

import (
	"globaldevtools.bbva.com/bitbucket/scm/nbdnt/nbdnt_gommander.git/gommander"
)

// This is an example main

func main() {
	config := gommander.Read("plan")
	config.Plan.Execute()
}
