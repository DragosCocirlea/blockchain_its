package main

import (
	"sync"

	protocol "github.com/libp2p/go-libp2p-protocol"
)

// protocol IDs
const (
	NodesProtocolID protocol.ID = "/bc_its/nodes"
)

// internal buffer for data received from region nodes
var speedReportsReceived map[GeoPoint]map[FloatString]float64
var usersReputationReceived map[string]float64
var alertsReceived map[Alert]AlertData

// global mutex used for accessing and modifying blockchain data
var blockchainMutex = sync.Mutex{}
var speedReportsMutex = sync.Mutex{}
var alertsMutex = sync.Mutex{}
var usersReputationMutex = sync.Mutex{}
var updateMutex = sync.Mutex{}
var updateCond = sync.NewCond(&updateMutex)

func initVariables() {
	speedReportsReceived = make(map[GeoPoint]map[FloatString]float64)
	usersReputationReceived = make(map[string]float64)
	alertsReceived = make(map[Alert]AlertData)
}
