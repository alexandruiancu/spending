package loader

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	resource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"spending/bldrec"
	"spending/common"

	capnp "capnproto.org/go/capnp/v3"
	zmq "github.com/pebbe/zmq4"
)

var glblMeterProvider *sdkmetric.MeterProvider

func initConn() (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient("localhost:4317",
		// Note the use of insecure transport here.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

func createDebitView() sdkmetric.View {
	// ---- provide view instead of instrument --------------------------------
	debitView := sdkmetric.NewView(
		sdkmetric.Instrument{
			Kind:        sdkmetric.InstrumentKindHistogram,
			Name:        "op.debit",
			Unit:        "ron",
			Description: "bank account debit in currency RON",
		},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: []float64{10, 50, 100, 500, 1000, 5000, 10000},
			},
		},
	)
	return debitView
}

func createResource(ctx context.Context) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("spending-loader"),
		semconv.ServiceVersion("0.1.0"),
	)

}

func CreateMetricsPipeline(ctx context.Context) error {

	conn, err := initConn()
	if err != nil {
		log.Fatalf("failed to initialize gRPC connection: %v", err)

		return nil
	}
	// ---- OTLP gRPC exporter ------------------------------------------------
	otlpExp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		log.Fatalf("failed to create OTLP exporter: %v", err)

		return nil
	}

	glblMeterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(otlpExp)),
		sdkmetric.WithResource(createResource(ctx)),
		sdkmetric.WithView(createDebitView()))
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
