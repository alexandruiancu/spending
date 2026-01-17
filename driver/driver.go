package main

import (
	"context"
	"fmt"
	"log"
	"me/bldrec"
	"me/common"
	"me/loader"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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

	args := os.Args[1:]
	config := common.ReadConfig(args[0])
	go bldrec.Process(config)
	go loader.StartLoadBalancer(config)

	<-ctx.Done()
	fmt.Println("Shutting down driver.")
}
