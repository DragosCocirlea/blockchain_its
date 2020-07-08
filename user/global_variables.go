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

var allAlerts map[Alert]AlertData
var alertsToDisplay []Alert

var currentLocation GeoPoint
var directionLat float64
var directionLong float64
var ticks int64
var currentTick int64

var locationMutex sync.Mutex
var alertsMutex sync.Mutex

func initVariables() {
	allAlerts = make(map[Alert]AlertData)
	locationMutex = sync.Mutex{}
	alertsMutex = sync.Mutex{}

	currentLocation = GeoPoint{45, 27}
	directionLat = 0.003
	directionLong = 0
	ticks = 1
	currentTick = 0
}
