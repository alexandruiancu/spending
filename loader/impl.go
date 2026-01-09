package loader

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"spending/bldrec"
	"spending/common"

	capnp "capnproto.org/go/capnp/v3"
	zmq "github.com/pebbe/zmq4"
)

var glblMeterProvider *sdkmetric.MeterProvider

// Initialize a gRPC connection to be used by both the tracer and meter
// providers.
func initConn() (*grpc.ClientConn, error) {
	// It connects the OpenTelemetry Collector through local gRPC connection.
	// You may replace `localhost:4317` with your endpoint.
	conn, err := grpc.NewClient("localhost:4317",
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

func CreateMetricsPipeline(ctx context.Context) error {

	comm, err := initConn()
	if err != nil {
		log.Fatalf("failed to initialize gRPC connection: %v", err)

		return nil
	}
	// ---- OTLP gRPC exporter -------------------------------------------------
	// Adjust the endpoint as needed (default: localhost:4317)
	otlpExp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(comm))
	if err != nil {
		log.Fatalf("failed to create OTLP exporter: %v", err)

		return nil
	}

	glblMeterProvider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(otlpExp)))
	otel.SetMeterProvider(glblMeterProvider)
	return nil
}

func CreateDebitInstrument() metric.Float64Histogram {

	// ---- Create the histogram instrument ------------------------------------
	meter := glblMeterProvider.Meter("spending/loader")
	hist, err := meter.Float64Histogram(
		"op.debit",
		metric.WithUnit("ron"),
		metric.WithDescription("bank account debit in currency RON"),
	)
	if err != nil {
		log.Fatalf("failed to create histogram: %v", err)
	}

	return hist
}

func ShutdownMetric(ctx context.Context) {
	if err := glblMeterProvider.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown MeterProvider: %s", err)
	}
}

func StartLoadBalancer() {
	frontend, _ := zmq.NewSocket(zmq.ROUTER)
	defer frontend.Close()

	config := common.ReadConfig("../config.txt")
	port := config["frontend_port"]
	frontend.Bind(fmt.Sprintf("tcp://localhost:%s", port))

	backend, _ := zmq.NewSocket(zmq.DEALER)
	defer backend.Close()
	backend.Bind("tcp://localhost:5556")
	for i := 0; i < 5; i++ {
		go startWorker(i)
	}

	zmq.Proxy(frontend, backend, nil)
}

func startWorker(id int) {
	socket, _ := zmq.NewSocket(zmq.REP)
	defer socket.Close()
	socket.Connect("tcp://localhost:5556")

	for {
		zmqMsgBytes, _ := socket.RecvBytes(0)
		// Wrap in a Cap’n Proto message (read‑only)
		msg, err := capnp.Unmarshal(zmqMsgBytes)
		if err != nil {
			log.Fatalf("capnp message: %v", err)
		}
		record, err := bldrec.ReadRootRecord(msg)
		if err != nil {
			log.Fatalf("read struct: %v", err)
		}
		desc, _ := record.SDescription()
		tmp, err := fmt.Printf("Worker %d received: %s\n", id, desc)
		if err != nil {
			continue
		}
		//println(desc)
		socket.Send(fmt.Sprintf("Reply from worker %d", tmp), 0)
	}
}
