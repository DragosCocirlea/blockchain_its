package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	// Parse options from the command line
	listenPort := flag.Int("port", 0, "wait for incoming connections")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	if *listenPort == 0 {
		log.Fatal("Please provide a port to listen on with -port")
	}

	// Make a host that listens on the given multiaddress
	ha, err := makeMasterHost(*listenPort, *seed)
	checkErrorFatal(err)

	initVariables()
	initBlockchain()
	go updateBlockchain()

	// Stream handler that opens when specified as a target by a region node
	ha.SetStreamHandler(NodesProtocolID, handleRegionNodeStream)

	// hang forever
	select {}
}

// create and append a new block based on the data waiting to be added to the blockchain
func updateBlockchain() {
	for {
		// wait a period of time
		time.Sleep(10 * time.Second)

		// block any transactions from being added
		speedReportsMutex.Lock()
		usersReputationMutex.Lock()
		alertsMutex.Lock()

		// only generate a new block if there are new transactions
		if len(speedReportsReceived) != 0 {
			updateCond.L.Lock()
			blockchainMutex.Lock()

			newBlock := generateBlock(ITSBlockchain[len(ITSBlockchain)-1])
			ITSBlockchain = append(ITSBlockchain, newBlock)

			bytes, err := json.MarshalIndent(newBlock, "", "  ")
			checkErrorFatal(err)
			fmt.Printf("\n\nNew block created:\n\x1b[32m%s\x1b[0m\n\n", string(bytes))

			// clear maps containing data awaiting to be added
			// optimized by the compiler: https://golang.org/cl/110055
			for k := range speedReportsReceived {
				delete(speedReportsReceived, k)
			}
			for k := range usersReputationReceived {
				delete(usersReputationReceived, k)
			}
			for k := range alertsReceived {
				delete(alertsReceived, k)
			}

			// let goroutines know that they can send the update to region nodes
			updateCond.Broadcast()
			updateCond.L.Unlock()

			blockchainMutex.Unlock()
		}

		// let new transactions be accepted
		speedReportsMutex.Unlock()
		usersReputationMutex.Unlock()
		alertsMutex.Unlock()
	}
}
