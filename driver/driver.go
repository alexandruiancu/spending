package main

import (
	"spending/bldrec"
	"spending/loader"
)

func main() {
	go bldrec.Process()
	go loader.StartLoadBalancer()
	select {}
}
