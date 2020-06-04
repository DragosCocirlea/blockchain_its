package main

func computeAlertUserReputation(alert Alert) {
	alertUsersAnswers := alertsUsersAnswers[alert]

	// compute consensus answer
	consensusAnswer := true
	if regionActiveAlerts[alert].Confirmations < regionActiveAlerts[alert].Denies {
		consensusAnswer = false
	}

	for userID, userAnswer := range alertUsersAnswers {

		userReputation, exists := allUserReputation[userID]
		if !exists {
			userReputation = 0.42
		}

		deltaRep := getDeltaRep(0.01, userReputation)

		if consensusAnswer == userAnswer {
			newUserReputation[userID] += deltaRep
		} else {
			newUserReputation[userID] -= deltaRep
		}
	}
}

func computeSpeedUsersReputation() {
	usersAnswersMutex.Lock()
	defer usersAnswersMutex.Unlock()

	// reputation derived from speed reports
	for userID, userAnswers := range usersSpeedAnswers {
		// repChange in [-1, +1]
		repChange := float64(userAnswers.Good-userAnswers.Bad) / float64(userAnswers.Good+userAnswers.Bad)

		// repChange in [-0.01, +0.01]
		repChange *= 0.01

		userReputation, exists := allUserReputation[userID]
		if !exists {
			userReputation = 0.42
		}

		deltaRep := getDeltaRep(repChange, userReputation)

		// save new user reputation
		newUserReputation[userID] += deltaRep
	}
}

func getDeltaRep(repChange float64, userReputation float64) float64 {
	if repChange > 0 && userReputation > 0.5 {
		return repChange * (1 - userReputation)
	} else {
		return repChange * userReputation
	}
}
