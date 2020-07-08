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
	way := flag.String("way", "", "")
	flag.Parse()

	if *listenPort == 0 {
		log.Fatal("Please provide a port to listen on with -port")
	}
	if *target == "" {
		log.Fatal("Please provide a region node to connect to with -peer")
	}

	initVariables()
	if *way != "" {
		directionLat *= -1
	}

	// Make a host that listens on the given multiaddress
	ha, err := makeUserHost(*listenPort, *target, *seed)
	checkErrorFatal(err)

	// Open new streams to region node
	speedStream := newSpeedStream(ha, *target)
	alertStream := newAlertStream(ha, *target)

	// Create a buffered stream so that read and writes are non blocking.
	regionSpeedRW := bufio.NewReadWriter(bufio.NewReader(speedStream), bufio.NewWriter(speedStream))
	regionAlertRW := bufio.NewReadWriter(bufio.NewReader(alertStream), bufio.NewWriter(alertStream))

	sendUserID(regionSpeedRW, ha.ID().Pretty())
	sendUserID(regionAlertRW, ha.ID().Pretty())

	// Create go routines that pass data between user node(this) and the region node
	go sendRegionSpeedMessage(regionSpeedRW)
	go sendRegionAlertMessage(regionAlertRW)
	go readRegionAlertMessage(regionAlertRW)

	// hang forever
	select {}
}

func checkErrorFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
