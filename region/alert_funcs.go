package main

import "time"

func parseRegionAlerts() {
	alertUserMessageMutex.Lock()
	defer alertUserMessageMutex.Unlock()

	currentTime := time.Now()

	// regionActiveAlerts also contain newUserAlerts
	for alert, status := range regionActiveAlerts {

		// only check the number of answers after 10 minutes from the alert creation
		if currentTime.Sub(regionActiveAlerts[alert].Creation).Minutes() > 10 {

			// if the alert hasn't yet been verified, mark it as verified and award reputation based on the consensus answer
			if !status.Verified {
				computeAlertUserReputation(alert)

				status.Verified = true
				regionActiveAlerts[alert] = status
				newUserAlerts[alert] = status
			}

			// if there are more denies than confirmations, deactivate the alert
			if status.Confirmations < status.Denies {
				deactivateAlert(alert, status)
				continue
			}
		}

		// if more than 1 hour passed since the latest confirmation, deactivate the alert
		if currentTime.Sub(status.LatestConfirmation).Hours() > 1 {
			deactivateAlert(alert, status)
		}
	}
}

func deactivateAlert(alert Alert, status AlertData) {
	// mark the alert as inactive
	status.Active = false

	// add it to the buffer
	newUserAlerts[alert] = status

	// remove alert from the internal data structure containing all active alerts
	delete(allActiveAlerts, alert)
	delete(regionActiveAlerts, alert)

	// remove alert from map containing users answers so that users can make a new such alert
	delete(alertsUsersAnswers, alert)
}
