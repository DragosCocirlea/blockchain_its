package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

func readInitialBlockchain(masterRW *bufio.ReadWriter) {
	fmt.Println("Waiting for blockchain from master node...")

	str, err := masterRW.ReadString('\n')
	checkErrorFatal(err)
	str = strings.Replace(str, "\n", "", -1)

	err = json.Unmarshal([]byte(str), &ITSBlockchain)
	checkErrorFatal(err)

	bytes, err := json.MarshalIndent(ITSBlockchain, "", "  ")
	checkErrorFatal(err)

	fmt.Println(Green("\n\nBlockchain received:"))
	fmt.Println(Green(string(bytes)))
}

func parseInitialBlockchain() {
	fmt.Println("Parsing the blockchain...")

	alertsAux := make(map[Alert]AlertData)

	// iterate in reverse order - this way only the latest data is taken into account and older data is simply skipped
	// don't check block 0 - Genesis block -> no data
	for i := len(ITSBlockchain) - 1; i > 0; i-- {
		// for each speed report, add it only if it doesn't exist
		for coord, speed := range ITSBlockchain[i].Data.SpeedReports {
			_, exists := allSpeedReports[coord]
			if !exists {
				allSpeedReports[coord] = speed
			}
		}

		// since the reputation data consists of deltas it all needs to be summed up
		for userID, repChange := range ITSBlockchain[i].Data.UsersReputation {
			allUserReputation[userID] += repChange
		}

		// for each alert, add it only if it doesn't exist, even if it gets deactivated
		for alert, data := range ITSBlockchain[i].Data.Alerts {
			_, exists := alertsAux[alert]
			if !exists {
				alertsAux[alert] = data
			}
		}
	}

	// add the initial reputation the result of the summed up delta reps
	for userID := range allUserReputation {
		allUserReputation[userID] += 0.42
	}

	// remove the alerts that have been deactivated
	for alert, data := range alertsAux {
		if data.Active {
			allActiveAlerts[alert] = data
		}
	}

	fmt.Printf("Parsing finished. You can now connect to this region node.\n\n")
}

func readBlockchainUpdate(masterRW *bufio.ReadWriter) {
	for {
		str, err := masterRW.ReadString('\n')
		checkErrorFatal(err)
		str = strings.Replace(str, "\n", "", -1)

		// unmarshall the new block
		newBlock := ITSBlock{}
		err = json.Unmarshal([]byte(str), &newBlock)
		checkErrorFatal(err)

		bytes, err := json.MarshalIndent(newBlock, "", "  ")
		checkErrorFatal(err)

		fmt.Println(Green("\nNew block received:"))
		fmt.Println(Green(string(bytes)))

		// try to append the newly created block to the internal blockchain
		blockchainMutex.Lock()
		propagateAlertCond.L.Lock()
		if isBlockValid(newBlock, ITSBlockchain[len(ITSBlockchain)-1]) {
			ITSBlockchain = append(ITSBlockchain, newBlock)

			// parse the new speeds
			for coord, speed := range newBlock.Data.SpeedReports {
				allSpeedReports[coord] = speed
			}

			// parse the new alerts - update / delete
			for alert, status := range newBlock.Data.Alerts {

				if status.Active {
					allActiveAlerts[alert] = status
				} else {
					delete(allActiveAlerts, alert)
				}
			}

			// parse the new reputation
			for id, rep := range newBlock.Data.UsersReputation {
				// if the user doesn't exist in the internal cache, assign it the initial rep value
				_, exists := allUserReputation[id]
				if !exists {
					allUserReputation[id] = 0.42
				}

				allUserReputation[id] += rep
			}

		} else {
			log.Fatal("New block append error") // extremely unlikely
		}
		propagateAlertCond.Broadcast()
		propagateAlertCond.L.Unlock()
		blockchainMutex.Unlock()
	}
}

func sendMasterTransaction(masterRW *bufio.ReadWriter) {
	for {
		time.Sleep(10 * time.Second)

		// only generate a new transaction if there are new speed messages from users
		if len(speedUserMessages) != 0 {
			parseUsersSpeedMessages()
			computeSpeedUsersReputation()
			parseRegionAlerts()

			// convert bearing type from float64 to FloatString
			newStringSpeedReports := make(map[GeoPoint]map[FloatString]float64)
			for geoPoint, reportsFloat := range newSpeedReports {
				reportsString := make(map[FloatString]float64)
				for bearing, speed := range reportsFloat {
					reportsString[FloatString(bearing)] = speed
				}

				newStringSpeedReports[geoPoint] = reportsString
			}

			// create and send the message to the master node
			message := MessageRegionToMaster{newStringSpeedReports, newUserAlerts, newUserReputation}
			bytes, err := json.Marshal(message)
			checkErrorFatal(err)
			stringToSend := string(bytes)
			masterRW.WriteString(fmt.Sprintf("%s\n", stringToSend))
			masterRW.Flush()

			// print what has been sent
			bytes, err = json.MarshalIndent(message, "", "  ")
			checkErrorFatal(err)
			fmt.Println(Yellow("\nSent transaction to master node:"))
			fmt.Println(Yellow(string(bytes)))

			// clear buffer containing data that needs to be sent to the master node
			for k := range newSpeedReports {
				delete(newSpeedReports, k)
			}
			for k := range newUserReputation {
				delete(newUserReputation, k)
			}
			for k := range newUserAlerts {
				delete(newUserAlerts, k)
			}

			for k := range newStringSpeedReports {
				delete(newStringSpeedReports, k)
			}
		}
	}
}
