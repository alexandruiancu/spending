package loader

import (
	"fmt"
	"log"

	"spending/bldrec"
	"spending/common"

	capnp "capnproto.org/go/capnp/v3"
	zmq "github.com/pebbe/zmq4"
)

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
