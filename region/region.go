package main

import (
	"bufio"
	"flag"
	"log"
)

func main() {
	// Parse options from the command line
	listenPort := flag.Int("port", 0, "wait for incoming connections")
	target := flag.String("peer", "", "target peer to dial")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	if *listenPort == 0 {
		log.Fatal("Please provide a port to listen on with -port")
	}
	if *target == "" {
		log.Fatal("Please provide a master node to connect to with -peer")
	}

	initVariables()

	// Make a host that listens on the given multiaddress
	ha, err := makeRegionHost(*listenPort, *seed)
	checkErrorFatal(err)

	// Stream handlers that open when specified as a target by a user
	ha.SetStreamHandler(UserSpeedProtocolID, handleUserSpeedStream)
	ha.SetStreamHandler(UserAlertProtocolID, handleUserAlertStream)

	// Open a new stream to master node
	stream := newStream(ha, *target)

	// Create a buffered stream so that read and writes are non blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	readInitialBlockchain(rw)
	parseInitialBlockchain()

	// Create go routines that pass data between region node(this) and the master node
	go readBlockchainUpdate(rw)
	go sendMasterTransaction(rw)

	// hang forever
	select {}
}
