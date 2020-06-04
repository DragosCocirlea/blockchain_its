package main

import (
	"sync"

	protocol "github.com/libp2p/go-libp2p-protocol"
)

// protocol IDs
const (
	NodesProtocolID     protocol.ID = "/bc_its/nodes"
	UserSpeedProtocolID protocol.ID = "/bc_its/user_speed"
	UserAlertProtocolID protocol.ID = "/bc_its/user_alert"
)

// internal data structures (easier data manipulation)
var allSpeedReports map[GeoPoint]map[FloatString]float64
var allUserReputation map[string]float64
var allActiveAlerts map[Alert]AlertData
var regionActiveAlerts map[Alert]AlertData

// internal buffer for data derived from messages received from users
var newSpeedReports map[GeoPoint]map[float64]float64
var newUserReputation map[string]float64
var newUserAlerts map[Alert]AlertData

// buffer for speed mesages from users
var speedUserMessages map[GeoPoint][]UserReport

// how many good/bag speed reports a user has made
var usersSpeedAnswers map[string]UserSpeedAnswers

// a map which contains the alerts and how to users voted for each of them
var alertsUsersAnswers map[Alert]map[string]bool

// map containing the last known location of a user
var userPosition map[string]GeoPoint

// global mutex used for accessing and modifying blockchain data
var blockchainMutex sync.Mutex
var speedUserMessageMutex sync.Mutex
var usersAnswersMutex sync.Mutex
var alertUserMessageMutex sync.Mutex
var userLocationMutexes map[string]*sync.Mutex
var propagateAlertMutex = sync.Mutex{}
var propagateAlertCond = sync.NewCond(&propagateAlertMutex)

func initVariables() {
	allSpeedReports = make(map[GeoPoint]map[FloatString]float64)
	allUserReputation = make(map[string]float64)
	allActiveAlerts = make(map[Alert]AlertData)
	regionActiveAlerts = make(map[Alert]AlertData)

	newSpeedReports = make(map[GeoPoint]map[float64]float64)
	newUserReputation = make(map[string]float64)
	newUserAlerts = make(map[Alert]AlertData)

	speedUserMessages = make(map[GeoPoint][]UserReport)
	usersSpeedAnswers = make(map[string]UserSpeedAnswers)
	alertsUsersAnswers = make(map[Alert]map[string]bool)
	userPosition = make(map[string]GeoPoint)

	blockchainMutex = sync.Mutex{}
	speedUserMessageMutex = sync.Mutex{}
	usersAnswersMutex = sync.Mutex{}
	alertUserMessageMutex = sync.Mutex{}
	userLocationMutexes = make(map[string]*sync.Mutex)
}
