package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func showAllAlerts() {
	alertsMutex.Lock()

	// check if there is any alert to show
	if len(allAlerts) == 0 {
		fmt.Println("There are no alerts to show")
		alertsMutex.Unlock()
		return
	}

	// get current location
	locationMutex.Lock()
	location := currentLocation
	locationMutex.Unlock()

	// gather all alerts in a slice and compute their distance from the users location
	alertsToDisplay = make([]Alert, 0)
	alertsDistance := make(map[Alert]float64)
	for alert := range allAlerts {
		alertsToDisplay = append(alertsToDisplay, alert)
		alertsDistance[alert] = location.distanceTo(alert.Coord)
	}

	// sort alerts by the distance
	sort.Slice(alertsToDisplay, func(i, j int) bool {
		return alertsDistance[alertsToDisplay[i]] < alertsDistance[alertsToDisplay[j]]
	})

	alertsMutex.Unlock()

	fmt.Println("Currently active alerts:")
	for i, alert := range alertsToDisplay {
		distance := alertsDistance[alert]
		distance /= 1000                          //to km
		distance = math.Round(distance*100) / 100 // round to 2 decimals

		fmt.Printf("\t%d. [%f, %f] "+alertIntToString(alert.AlertType)+" - %f km away\n", i, alert.Coord.Lat, alert.Coord.Long, distance)
	}
}

func chooseExistingAlert(stdReader *bufio.Reader) (AlertMessageUserToRegion, error) {
	// check if the user has actually seen any alerts
	if len(alertsToDisplay) == 0 {
		fmt.Println("You haven't seen any alerts yet so you can't give any feedback")
		err := errors.New("No alerts displayed yet")
		return AlertMessageUserToRegion{}, err
	}

	// get alert
	fmt.Println("Which alert would you like to give feedback to?")
	fmt.Printf("\t> ")

	alertIndexInput, err := stdReader.ReadString('\n')
	checkErrorFatal(err)
	alertIndexInput = strings.Replace(alertIndexInput, "\n", "", -1)

	alertIndex, err := strconv.Atoi(alertIndexInput)
	if err != nil {
		fmt.Println("Please insert an integer next time")
		return AlertMessageUserToRegion{}, err
	}

	if alertIndex < 0 || alertIndex >= len(alertsToDisplay) {
		fmt.Println("Please select a valid index next time")
		err := errors.New("Invalid error index")
		return AlertMessageUserToRegion{}, err
	}

	alert := alertsToDisplay[alertIndex]

	// check distance
	locationMutex.Lock()
	location := currentLocation
	locationMutex.Unlock()

	distance := location.distanceTo(alert.Coord)
	if distance > 1000 {
		fmt.Println("The distance between you and the alert cannot be greated than 1km")
		err := errors.New("Alert too far away")
		return AlertMessageUserToRegion{}, err
	}

	// check if the alert is still active
	var alertActive bool

	fmt.Println("Is the alert still valid?(y/n)")
	fmt.Printf("\t> ")
	alertActiveInput, err := stdReader.ReadString('\n')
	checkErrorFatal(err)
	alertActiveInput = strings.Replace(alertActiveInput, "\n", "", -1)

	if alertActiveInput == "y" {
		alertActive = true
	} else if alertActiveInput == "n" {
		alertActive = false
	} else {
		fmt.Println("Please type 'y' or 'n' when saying whether the alert is still active")
		err := errors.New("Invalid error input - active")
		return AlertMessageUserToRegion{}, err
	}

	// return the message to be sent
	return AlertMessageUserToRegion{alert.Coord, alert.AlertType, alertActive}, nil
}

func createNewAlert(stdReader *bufio.Reader) (AlertMessageUserToRegion, error) {
	// get current location
	locationMutex.Lock()
	location := currentLocation
	locationMutex.Unlock()

	fmt.Println("What kind of alert is it?")
	fmt.Println("\t1 - Car crash")
	fmt.Println("\t2 - Pothole")
	fmt.Println("\t3 - Roadkill")
	fmt.Println("\tother integer - Basic alert")
	fmt.Printf("\t> ")

	alertTypeInput, err := stdReader.ReadString('\n')
	checkErrorFatal(err)
	alertTypeInput = strings.Replace(alertTypeInput, "\n", "", -1)

	alertType, err := strconv.Atoi(alertTypeInput)
	if err != nil {
		fmt.Println("Please input a valid alert type")
		return AlertMessageUserToRegion{}, err
	}

	// return the message to be sent
	return AlertMessageUserToRegion{location, alertType, true}, nil
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
