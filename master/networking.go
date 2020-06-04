package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"

	net "github.com/libp2p/go-libp2p-net"
)

func handleRegionNodeStream(s net.Stream) {
	log.Println("Got a new region node stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	sendBlockchain(rw)

	// Create goroutines that continuously pass data between master node(this) and region node
	go readRegionTransaction(rw)
	go sendBlockchainUpdate(rw)
}

// send the blockchain to a newly connected region node
func sendBlockchain(rw *bufio.ReadWriter) {
	blockchainMutex.Lock()
	bytes, err := json.Marshal(ITSBlockchain)
	checkErrorFatal(err)
	blockchainMutex.Unlock()

	rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	rw.Flush()
}
