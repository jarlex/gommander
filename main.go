package main

import (
	"globaldevtools.bbva.com/bitbucket/scm/nbdnt/nbdnt_gommander.git/gommander"
)

func main() {
	//gommander.Execute()
	config := gommander.Read("/home/aruiz/go/src/globaldevtools.bbva.com/bitbucket/scm/nhedd/nhedd_chameleon_pki_service.git/gommander")
	config.Plan.Execute()
}
