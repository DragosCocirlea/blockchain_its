package main

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"time"
)

// ITSBlockchain is a series of validated Blocks
var ITSBlockchain []ITSBlock

// ITSBlockData - block data datatype
type ITSBlockData struct {
	SpeedReports    map[GeoPoint]map[FloatString]float64
	Alerts          map[Alert]AlertData
	UsersReputation map[string]float64
}

// ITSBlock - each 'item' in the blockchain
type ITSBlock struct {
	Index     int
	Timestamp string
	Data      ITSBlockData
	Hash      string
	PrevHash  string
}

// block initialization.
func initBlockchain() {
	t := time.Now()
	genesisBlock := ITSBlock{}
	genesisBlock = ITSBlock{0, t.String(), ITSBlockData{}, calculateHash(genesisBlock), ""}
	ITSBlockchain = append(ITSBlockchain, genesisBlock)
}

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock ITSBlock) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hashing
func calculateHash(block ITSBlock) string {
	// block to string
	record := strconv.Itoa(block.Index) + block.Timestamp + block.PrevHash

	var speedReportsString, alertsString, reputationChangesString string

	var wg sync.WaitGroup
	wg.Add(3)
	go speedReportsWorker(&wg, &speedReportsString, block.Data.SpeedReports)
	go alertsWorker(&wg, &alertsString, block.Data.Alerts)
	go reputationChangesWorker(&wg, &reputationChangesString, block.Data.UsersReputation)
	wg.Wait()

	record += speedReportsString + alertsString + reputationChangesString

	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)

	return hex.EncodeToString(hashed)
}

func speedReportsWorker(wg *sync.WaitGroup, str *string, blockSpeedReports map[GeoPoint]map[FloatString]float64) {
	defer wg.Done()

	// Collect all speed reports coords
	coords := make([]GeoPoint, 0)
	for k := range blockSpeedReports {
		coords = append(coords, k)
	}

	// sort the speed report by coords
	sort.Slice(coords, func(i, j int) bool {
		if coords[i].Lat != coords[j].Lat {
			return coords[i].Lat < coords[j].Lat
		}
		return coords[i].Long < coords[j].Long
	})

	// iterate the map by key (Go maps do not maintain the insertion order)
	for _, coord := range coords {
		speeds := blockSpeedReports[coord]
		*str = *str + floatToString(coord.Lat) + " " + floatToString(coord.Long) + ": "

		// collect bearings in this geopoint
		bearings := make([]float64, 0)
		for k := range speeds {
			bearings = append(bearings, float64(k))
		}
		// sort
		sort.Float64s(bearings)

		// iterate
		for _, bearing := range bearings {
			*str = *str + floatToString(bearing) + " - " + floatToString(speeds[FloatString(bearing)]) + ". "
		}
	}
}

func alertsWorker(wg *sync.WaitGroup, str *string, blockAlerts map[Alert]AlertData) {
	defer wg.Done()

	// Collect all alert keys
	alerts := make([]Alert, 0)
	for k := range blockAlerts {
		alerts = append(alerts, k)
	}
	// sort
	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].Coord.Lat != alerts[j].Coord.Lat {
			return alerts[i].Coord.Lat < alerts[j].Coord.Lat
		}
		if alerts[i].Coord.Long != alerts[j].Coord.Long {
			return alerts[i].Coord.Long < alerts[j].Coord.Long
		}
		return alerts[i].AlertType < alerts[j].AlertType
	})
	// iterate
	for _, alert := range alerts {
		ans := blockAlerts[alert]
		*str = *str + floatToString(alert.Coord.Lat) + " " + floatToString(alert.Coord.Long) + " " + strconv.Itoa(alert.AlertType) + " - " +
			strconv.FormatBool(ans.Active) + ", " + strconv.FormatBool(ans.Verified) + ", " +
			floatToString(ans.Confirmations) + ", " + floatToString(ans.Denies) + ", " +
			ans.Creation.String() + ", " + ans.LatestConfirmation.String() + "; "
	}
}

func reputationChangesWorker(wg *sync.WaitGroup, str *string, blockReputationChanges map[string]float64) {
	defer wg.Done()

	// Collect all reputation user ids and sort them
	ids := make([]string, 0)
	for k := range blockReputationChanges {
		ids = append(ids, k)
	}
	//sort
	sort.Strings(ids)
	//iterate
	for _, id := range ids {
		rep := blockReputationChanges[id]
		*str = *str + "{" + id + "} - " + floatToString(rep) + ", "
	}
}

// create a new block using previous block's hash
func generateBlock(oldBlock ITSBlock) ITSBlock {
	var newBlock ITSBlock
	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()

	newBlock.Data.SpeedReports = make(map[GeoPoint]map[FloatString]float64)
	for k, v := range speedReportsReceived {
		newBlock.Data.SpeedReports[k] = v
	}

	newBlock.Data.UsersReputation = make(map[string]float64)
	for k, v := range usersReputationReceived {
		newBlock.Data.UsersReputation[k] = v
	}

	newBlock.Data.Alerts = make(map[Alert]AlertData)
	for k, v := range alertsReceived {
		newBlock.Data.Alerts[k] = v
	}

	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock
}

func floatToString(inputNum float64) string {
	return strconv.FormatFloat(inputNum, 'f', 6, 64)
}
