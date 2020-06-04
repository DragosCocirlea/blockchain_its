package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
)

// message received from another node (blockchain update)
func readRegionTransaction(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		checkErrorFatal(err)
		str = strings.Replace(str, "\n", "", -1)

		// unmarshall the message from the region node
		newMessage := MessageRegionToMaster{}
		err = json.Unmarshal([]byte(str), &newMessage)
		checkErrorFatal(err)

		// print what has been received
		bytes, err := json.MarshalIndent(newMessage, "", "  ")
		checkErrorFatal(err)
		fmt.Println("Received transaction from a region node:")
		fmt.Println(Yellow(string(bytes)))
		fmt.Println()

		// add the new speed reports to the pool of speed reports waiting to be added to the blockchain
		speedReportsMutex.Lock()
		for point, speed := range newMessage.SpeedReports {
			speedReportsReceived[point] = speed
		}
		speedReportsMutex.Unlock()

		// add the new user reputations to the pool of user reputations waiting to be added to the blockchain
		usersReputationMutex.Lock()
		for id, rep := range newMessage.UsersReputation {
			usersReputationReceived[id] = rep
		}
		usersReputationMutex.Unlock()

		// parse the new alerts and add them to the pool of new alerts waiting to be added to the blockchain
		alertsMutex.Lock()
		for alert, status := range newMessage.NewAlerts {
			alertsReceived[alert] = status
		}
		alertsMutex.Unlock()
	}
}

// resend the local blockchain to the other nodes after a new block has been created
func sendBlockchainUpdate(rw *bufio.ReadWriter) {
	for {
		// wait for signal that the blockchain has been updated
		updateCond.L.Lock()
		updateCond.Wait()
		updateCond.L.Unlock()

		// send last block created
		blockchainMutex.Lock()
		bytes, err := json.Marshal(ITSBlockchain[len(ITSBlockchain)-1])
		checkErrorFatal(err)
		blockchainMutex.Unlock()

		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
	}
}
