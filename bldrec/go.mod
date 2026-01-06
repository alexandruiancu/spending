module spending/bldrec

go 1.25.5

replace spending/common => ../common

require (
	capnproto.org/go/capnp/v3 v3.1.0-alpha.2
	spending/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/colega/zeropool v0.0.0-20230505084239-6fb4a4f75381 // indirect
	github.com/pebbe/zmq4 v1.4.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
)
