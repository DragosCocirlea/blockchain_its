package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func readUserID(userRW *bufio.ReadWriter) string {
	userID, err := userRW.ReadString('\n')
	checkErrorFatal(err)
	userID = strings.Replace(userID, "\n", "", -1)

	return userID
}

func readUserSpeedMessage(userRW *bufio.ReadWriter, userID string) {
	for {
		str, err := userRW.ReadString('\n')
		checkErrorFatal(err)
		str = strings.Replace(str, "\n", "", -1)

		// unmarshall the message
		newMessage := SpeedMessageUserToRegion{}
		err = json.Unmarshal([]byte(str), &newMessage)
		checkErrorFatal(err)

		// print report
		fmt.Println(Teal("\nNew speed report from " + userID + ":"))
		fmt.Println(Teal("\tcoords  - ", newMessage.Coord))
		fmt.Println(Teal("\tspeed   - ", newMessage.Speed))

		// update users location to the system and save the last known location
		userLocationMutexes[userID].Lock()
		lastKnownLocation, locationExists := userPosition[userID]
		userPosition[userID] = newMessage.Coord
		userLocationMutexes[userID].Unlock()

		// if this is the first speed report of this user, the bearing cannot be computed
		if !locationExists {
			fmt.Println(Teal("First speed report -> can't compute bearing"))
			continue
		}

		// if the position hasn't changed, ignore it
		if lastKnownLocation == newMessage.Coord {
			continue
		}

		// calculate bearing based on the two positions
		bearing := lastKnownLocation.bearingTo(newMessage.Coord)
		fmt.Println(Teal("\tbearing - ", bearing))

		// variable which clumps together the userID and the reported speed and bearing
		newUserReport := UserReport{newMessage.Speed, bearing, userID}

		// add the new speed report to the pool of speed reports waiting to be parsed
		speedUserMessageMutex.Lock()
		speedUserMessages[newMessage.Coord] = append(speedUserMessages[newMessage.Coord], newUserReport)
		speedUserMessageMutex.Unlock()
	}
}

func sendUserAlertMessage(userRW *bufio.ReadWriter, userID string) {
	for {
		// wait for signal that the region node has received a new block from the master node
		propagateAlertCond.L.Lock()
		propagateAlertCond.Wait()
		propagateAlertCond.L.Unlock()

		if len(ITSBlockchain[len(ITSBlockchain)-1].Data.Alerts) == 0 {
			continue
		}

		// send the latest alert updates to the user from the last block received
		bytes, err := json.Marshal(ITSBlockchain[len(ITSBlockchain)-1].Data.Alerts)
		checkErrorFatal(err)

		userRW.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		userRW.Flush()
	}
}

func readUserAlertMessage(userRW *bufio.ReadWriter, userID string) {
	for {
		str, err := userRW.ReadString('\n')
		checkErrorFatal(err)
		str = strings.Replace(str, "\n", "", -1)

		// unmarshall the message
		userAlertMsg := AlertMessageUserToRegion{}
		err = json.Unmarshal([]byte(str), &userAlertMsg)
		checkErrorFatal(err)

		// print report
		fmt.Println(Blue("\nNew alert from " + userID + ":"))
		fmt.Println(Blue("\tcoords - ", userAlertMsg.Coord))
		fmt.Println(Blue("\ttype   - ", alertIntToString(userAlertMsg.AlertType)))
		fmt.Println(Blue("\tstatus - ", userAlertMsg.Active))

		// idealy, if the alert is in a neighboring region, route it there:
		//		- either through the master node
		//		- or directly to the neighbouring region node
		// this way all alerts past this point are certain to be part of this region

		// get users last known position
		userLocationMutexes[userID].Lock()
		lastKnownLocation, locationExists := userPosition[userID]
		userLocationMutexes[userID].Unlock()

		// if the user sent a report without ever sending its location, skip the alert
		if !locationExists {
			fmt.Println(Blue("[ERROR] user location unknown"))
			continue
		}

		// if the user is not within a certain radius (1km) from the location of the alert, skip it
		distance := lastKnownLocation.distanceTo(userAlertMsg.Coord)
		if distance > 1000 {
			fmt.Println(Blue("[WARNING] distance would have been too big"))
			// continue
		}

		// create new alert key
		newAlert := Alert{userAlertMsg.Coord, userAlertMsg.AlertType}

		// check if the user has already responded to this alert
		userAlreadyAnswered := false
		alertAnswers, alertHasAnswers := alertsUsersAnswers[newAlert]
		if alertHasAnswers {
			_, userAlreadyAnswered = alertAnswers[userID]
		}
		if userAlreadyAnswered {
			fmt.Println(Blue("[ERROR] user has already answered this alert"))
			continue
		}

		alertUserMessageMutex.Lock()

		// check if the alert exists in the system
		alertData, existsInSystem := regionActiveAlerts[newAlert]

		// the user has denied an alert that doesn't exist in the system -> skip
		if !existsInSystem && !userAlertMsg.Active {
			alertUserMessageMutex.Unlock()
			continue
		}

		// get the users reputation
		userReputation, reputationExists := allUserReputation[userID]
		if !reputationExists {
			userReputation = 0.42
		}

		// get current time
		currentTime := time.Now()

		if !existsInSystem && userAlertMsg.Active {
			// the alert doesn't exist in the system and the user reports the alert as existing -> new alert
			alertData = AlertData{}

			alertData.Confirmations = userReputation
			alertData.Denies = 0
			alertData.Creation = currentTime
			alertData.LatestConfirmation = currentTime
			alertData.Verified = false
			alertData.Active = true

		} else if existsInSystem && userAlertMsg.Active {
			// the alert exists in the system and the user reports the alert as existing -> confirm
			alertData.Confirmations += userReputation

			// update the "latestConfirmation" field
			alertData.LatestConfirmation = currentTime

		} else if existsInSystem && !userAlertMsg.Active {
			// the alert exists in the system and the user reports the alert as not existing -> deny
			alertData.Denies += userReputation
		}

		// mark the fact that the user has responded to this alert
		if !alertHasAnswers {
			alertsUsersAnswers[newAlert] = map[string]bool{}
		}
		alertsUsersAnswers[newAlert][userID] = userAlertMsg.Active

		// save the new alert data
		regionActiveAlerts[newAlert] = alertData

		// add the new data to the buffer containing alerts that will be added to the blockchain
		newUserAlerts[newAlert] = alertData

		alertUserMessageMutex.Unlock()
	}
}

func alertIntToString(alertType int) string {
	switch alertType {
	case 1:
		return "Car crash"
	case 2:
		return "Pothole"
	case 3:
		return "Roadkill"
	default:
		return "Basic alert"
	}
}
