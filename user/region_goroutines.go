package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func sendUserID(regionRW *bufio.ReadWriter, userID string) {
	regionRW.WriteString(fmt.Sprintf("%s\n", userID))
	regionRW.Flush()
}

func sendRegionSpeedMessage(regionSpeedRW *bufio.ReadWriter) {
	for {
		time.Sleep(10 * time.Second)

		// create random speed between 90 and 110
		rand.Seed(time.Now().UnixNano())
		min := -10
		max := 10
		delta := rand.Intn(max-min+1) + min
		speed := 100 + delta

		// get and update current location
		locationMutex.Lock()
		speedMsg := SpeedMessageUserToRegion{currentLocation, float64(speed)}

		currentLocation.Lat += directionLat
		currentLocation.Long += directionLong

		currentTick = (currentTick + 1) % ticks
		if currentTick == 0 {
			directionLat *= -1
			directionLong *= -1
		}
		locationMutex.Unlock()

		// create and send the message
		bytes, err := json.Marshal(speedMsg)
		checkErrorFatal(err)

		regionSpeedRW.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		regionSpeedRW.Flush()
	}
}

func readRegionAlertMessage(regionAlertRW *bufio.ReadWriter) {
	for {
		str, err := regionAlertRW.ReadString('\n')
		checkErrorFatal(err)
		str = strings.Replace(str, "\n", "", -1)

		// unmarshall the message
		newAlerts := make(map[Alert]AlertData)
		err = json.Unmarshal([]byte(str), &newAlerts)
		if err != nil {
			fmt.Println("Am primit un kkt:", str)
			continue
		}
		checkErrorFatal(err)

		alertsMutex.Lock()

		// update the internal pool of alerts
		for alert, data := range newAlerts {
			if !data.Active {
				// if an alert has been marked as inactive, remove it
				delete(allAlerts, alert)
			} else {
				// else insert/update it
				allAlerts[alert] = data
			}
		}

		alertsMutex.Unlock()
	}
}

func sendRegionAlertMessage(regionAlertRW *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n==========================================================================")
		fmt.Println("What would you like to do?")
		fmt.Println("\t1 - See all alerts")
		fmt.Println("\t2 - Give feedback to an alert(note: distance has to be smaller than 1km)")
		fmt.Println("\t3 - Create a new alert")
		fmt.Println("\tr - Reset position")
		fmt.Printf("\t> ")

		userChoice, err := stdReader.ReadString('\n')
		checkErrorFatal(err)
		userChoice = strings.Replace(userChoice, "\n", "", -1)
		fmt.Println()

		if userChoice == "1" {
			showAllAlerts()
			continue
		}

		if userChoice == "r" {
			resetUserPosition(stdReader)
			continue
		}

		var alertMessage AlertMessageUserToRegion

		// create the message that needs to be sent
		if userChoice == "2" {
			alertMessage, err = chooseExistingAlert(stdReader)
		} else if userChoice == "3" {
			alertMessage, err = createNewAlert(stdReader)
		} else {
			fmt.Println("Please select a valid option.")
			continue
		}

		// if there was an error along the way, stop
		if err != nil {
			continue
		}

		// send the message
		bytes, err := json.Marshal(alertMessage)
		checkErrorFatal(err)
		regionAlertRW.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		regionAlertRW.Flush()

		fmt.Println("\nYour alert has been sent. Thank you!")
	}
}

func resetUserPosition(stdReader *bufio.Reader) {
	fmt.Println("\nWhere would you like to be positioned / headed?")

	// read Lat
	fmt.Printf("\tlat: ")
	latString, err := stdReader.ReadString('\n')
	if err != nil {
		fmt.Println("error reading lat")
		return
	}
	latString = strings.Replace(latString, "\n", "", -1)
	newLat, err := strconv.ParseFloat(latString, 64)
	if err != nil {
		fmt.Println("error converting lat to float")
		return
	}

	// read Long
	fmt.Printf("\tlong: ")
	longString, err := stdReader.ReadString('\n')
	if err != nil {
		fmt.Println("error reading long")
		return
	}
	longString = strings.Replace(longString, "\n", "", -1)
	newLong, err := strconv.ParseFloat(longString, 64)
	if err != nil {
		fmt.Println("error converting long to float")
		return
	}

	// read direction
	fmt.Printf("\tdirection(n, s, e, w): ")
	directionString, err := stdReader.ReadString('\n')
	if err != nil {
		fmt.Println("error reading direction")
		return
	}
	directionString = strings.Replace(directionString, "\n", "", -1)

	// read ticks
	fmt.Printf("\tticks: ")
	ticksString, err := stdReader.ReadString('\n')
	if err != nil {
		fmt.Println("error reading ticks")
		return
	}
	ticksString = strings.Replace(ticksString, "\n", "", -1)
	newTicks, err := strconv.ParseInt(ticksString, 10, 16)
	if err != nil {
		fmt.Println("error converting ticks to int")
		return
	}

	// update the users location
	locationMutex.Lock()

	currentLocation = GeoPoint{newLat, newLong}
	ticks = newTicks

	directionLat = 0
	directionLong = 0
	if strings.Contains(directionString, "n") {
		directionLat = 1
	}
	if strings.Contains(directionString, "s") {
		directionLat = -1
	}
	if strings.Contains(directionString, "e") {
		directionLong = 1
	}
	if strings.Contains(directionString, "w") {
		directionLong = -1
	}

	locationMutex.Unlock()
}
