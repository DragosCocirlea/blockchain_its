package main

import (
	"math"
	"sort"
)

func parseUsersSpeedMessages() {
	speedUserMessageMutex.Lock()
	defer speedUserMessageMutex.Unlock()

	for geoPoint, messages := range speedUserMessages {
		// compute the directions based on previous position
		mainBearings := computeMainBearings(messages)
		avgSpeeds := computeAverageSpeeds(messages, mainBearings)
		newSpeedReports[geoPoint] = avgSpeeds
		checkUsersSpeedMessages(avgSpeeds, messages)
	}

	// clear all speed user messages
	for k := range speedUserMessages {
		delete(speedUserMessages, k)
	}
}

var maxCircleSectorAngle = 25.0

func bearingsSmallestDistance(x, y float64) float64 {
	d1 := math.Abs(x - y) // doesn't pass 0
	d2 := 360 - d1        // passes 0
	return math.Min(d1, d2)
}

func bearingDirectionDistance(x, y float64, clockwise bool) float64 {
	var d1 float64
	var d2 float64

	if x <= y {
		d1 = y - x
		d2 = 360.0 - d1
	} else {
		d2 = x - y
		d1 = 360.0 - d2
	}

	if clockwise {
		return d1
	}

	return d2
}

func computeMainBearings(messages []UserReport) []float64 {
	allBearings := make([]float64, 0)
	mainBearings := make([]float64, 0)

	// gather all bearings
	for _, report := range messages {
		allBearings = append(allBearings, report.Bearing)
	}

	// sort the bearings
	sort.Float64s(allBearings)

	// find the start of the first sector depictiong a general direction
	firstIndex := findFirstBearingIndex(allBearings)

	firstSectorBearing := allBearings[firstIndex]
	distanceSum := 0.0
	bearingsDenominator := 1
	for i := (firstIndex + 1) % len(allBearings); ; i = (i + 1) % len(allBearings) {

		// if we have reached the firstIndex, compute the average bearing for the last sector
		if i == firstIndex {
			newMainBearing := firstSectorBearing + distanceSum/float64(bearingsDenominator)
			newMainBearing = math.Mod(newMainBearing, 360.0)
			mainBearings = append(mainBearings, newMainBearing)

			break
		}

		// clockwise distance between sector start and this bearing
		clockwiseDistance := bearingDirectionDistance(firstSectorBearing, allBearings[i], true)

		// if a new sector starts, compute the average bearing for the latest sector
		if clockwiseDistance > maxCircleSectorAngle {
			newMainBearing := firstSectorBearing + distanceSum/float64(bearingsDenominator)
			newMainBearing = math.Mod(newMainBearing, 360.0)
			mainBearings = append(mainBearings, newMainBearing)

			// mark the start of a new sector
			firstSectorBearing = allBearings[i]
			distanceSum = 0.0
			bearingsDenominator = 1

			continue
		}

		distanceSum += clockwiseDistance
		bearingsDenominator++
	}

	return mainBearings
}

func findFirstBearingIndex(bearings []float64) int {
	firstBearing := bearings[0]
	lastBearing := bearings[len(bearings)-1]

	// if the anti-clockwise gap between the first and the last bearings is bigger than maxCircleSectorAngle
	// it means that the first bearing is also a first bearing in a certain direction sector
	if bearingDirectionDistance(firstBearing, lastBearing, false) > maxCircleSectorAngle {
		return 0
	}

	// else check until finding a gap bigger than maxCircleSectorAngle
	for i, bearing := range bearings {
		if bearings[i+1]-bearing > maxCircleSectorAngle {
			return i + 1
		}
	}

	return 0
}

func computeAverageSpeeds(messages []UserReport, mainBearings []float64) map[float64]float64 {
	speedSums := make(map[float64]float64)
	reputationSums := make(map[float64]float64)
	avgSpeeds := make(map[float64]float64)

	// compute speed sums and reputation sums for each bearing
	for _, msg := range messages {
		assignedBearing := closestBearing(msg.Bearing, mainBearings)

		userReputation, exists := allUserReputation[msg.UserID]
		if !exists {
			userReputation = 0.42
		}

		speedSums[assignedBearing] += msg.Speed * userReputation
		reputationSums[assignedBearing] += userReputation
	}

	// compute the average speed for each bearing
	for _, bearing := range mainBearings {
		avgSpeed := speedSums[bearing] / reputationSums[bearing]
		avgSpeed = math.Round(avgSpeed*100) / 100

		avgSpeeds[bearing] = avgSpeed
	}

	return avgSpeeds
}

func closestBearing(bearing float64, mainBearings []float64) float64 {
	minDist := 361.0
	mainBearing := 0.0

	for _, possibleMainBearing := range mainBearings {
		dist := bearingsSmallestDistance(possibleMainBearing, bearing)

		if dist < minDist {
			minDist = dist
			mainBearing = possibleMainBearing
		}
	}

	return mainBearing
}

func checkUsersSpeedMessages(avgSpeeds map[float64]float64, messages []UserReport) {
	usersAnswersMutex.Lock()
	defer usersAnswersMutex.Unlock()

	// gather all possible bearings
	possibleBearings := make([]float64, 0)
	for bearing := range avgSpeeds {
		possibleBearings = append(possibleBearings, bearing)
	}

	for _, msg := range messages {
		// get the computed average speed of the corresponding user bearing
		assignedBearing := closestBearing(msg.Bearing, possibleBearings)
		avgSpeed := avgSpeeds[assignedBearing]

		// difference between avgSpeed and the speed reported by the user
		relativeDifference := math.Abs(1.0 - msg.Speed/avgSpeed)

		// get object that keeps count of the user's answers
		userAnswers, exists := usersSpeedAnswers[msg.UserID]
		if !exists {
			userAnswers = UserSpeedAnswers{0, 0}
		}

		if relativeDifference > 0.2 {
			userAnswers.Bad++
		} else {
			userAnswers.Good++
		}

		usersSpeedAnswers[msg.UserID] = userAnswers
	}
}
