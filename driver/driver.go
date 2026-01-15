package main

import (
	"fmt"
	"log"

	//"context"
	//"os"
	//"os/signal"

	"context"
	"os"
	"os/signal"
	"spending/bldrec"
	"spending/loader"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	//TODO: pass actual resource
	err := loader.CreateMetricsPipeline(ctx)
	if err != nil {
		log.Fatalf("error creating metric pipeline: %v", err)
	}
	defer loader.ShutdownMetric(ctx)
	fmt.Printf("Metrics pipeline created ...\n")
	err = loader.CreateDebitInstrument()
	if err != nil {
		log.Fatalf("error creating debit instrument: %v", err)
	}
	fmt.Printf("Debit instrument created ...\n")

	go bldrec.Process()
	go loader.StartLoadBalancer()
	select {}
}
