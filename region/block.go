package main

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
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

	// Collect all speed reports coords
	coords := make([]GeoPoint, 0)
	for k := range block.Data.SpeedReports {
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
		speeds := block.Data.SpeedReports[coord]
		record = record + floatToString(coord.Lat) + " " + floatToString(coord.Long) + ": "

		// collect bearings in this geopoint
		bearings := make([]float64, 0)
		for k := range speeds {
			bearings = append(bearings, float64(k))
		}
		// sort
		sort.Float64s(bearings)
		// iterate
		for _, bearing := range bearings {
			record = record + floatToString(bearing) + " - " + floatToString(speeds[FloatString(bearing)]) + ". "
		}
	}

	// Collect all alert keys
	alerts := make([]Alert, 0)
	for k := range block.Data.Alerts {
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
		ans := block.Data.Alerts[alert]
		record = record + floatToString(alert.Coord.Lat) + " " + floatToString(alert.Coord.Long) + " " + strconv.Itoa(alert.AlertType) + " - " +
			strconv.FormatBool(ans.Active) + ", " + strconv.FormatBool(ans.Verified) + ", " +
			floatToString(ans.Confirmations) + ", " + floatToString(ans.Denies) + ", " +
			ans.Creation.String() + ", " + ans.LatestConfirmation.String() + "; "
	}

	// Collect all reputation user ids and sort them
	ids := make([]string, 0)
	for k := range block.Data.UsersReputation {
		ids = append(ids, k)
	}
	//sort
	sort.Strings(ids)
	//iterate
	for _, id := range ids {
		rep := block.Data.UsersReputation[id]
		record = record + "{" + id + "} - " + floatToString(rep) + ", "
	}

	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)

	return hex.EncodeToString(hashed)
}

func floatToString(inputNum float64) string {
	return strconv.FormatFloat(inputNum, 'f', 6, 64)
}
