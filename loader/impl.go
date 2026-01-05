package loader

import (
	"fmt"
	"log"

	"spending/common"

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
	socket, _ := zmq.NewSocket(zmq.REQ)
	defer socket.Close()
	socket.Connect("tcp://localhost:5556")

	for {
		msg, _ := socket.Recv(0)
		log.Printf("Worker %d received: %s\n", id, msg)
		socket.Send(fmt.Sprintf("Reply from worker %d", id), 0)
	}
}
