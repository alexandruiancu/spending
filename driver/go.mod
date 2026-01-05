go 1.25.5

replace spending/bldrec => ../bldrec

require (
	spending/bldrec v0.0.0-00010101000000-000000000000
	spending/loader v0.0.0-00010101000000-000000000000
)

require (
	github.com/pebbe/zmq4 v1.4.0 // indirect
	spending/common v0.0.0-00010101000000-000000000000 // indirect
)

replace spending/common => ../common

replace spending/loader => ../loader

module spending/driver
