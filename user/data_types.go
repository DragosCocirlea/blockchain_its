package main

import (
	"encoding/json"
	"math"
	"time"
)

// SpeedMessageUserToRegion - message the user node send to the region node
type SpeedMessageUserToRegion struct {
	Coord GeoPoint
	Speed float64
}

// AlertMessageUserToRegion - alert message from user to region node
type AlertMessageUserToRegion struct {
	Coord     GeoPoint
	AlertType int
	Active    bool
}

// GeoPoint - point in geographic coordinate system
type GeoPoint struct {
	Lat  float64
	Long float64
}

func (g GeoPoint) MarshalText() (text []byte, err error) {
	type x GeoPoint
	return json.Marshal(x(g))
}

func (g *GeoPoint) UnmarshalText(text []byte) error {
	type x GeoPoint
	return json.Unmarshal(text, (*x)(g))
}

// harvesine formula taken from: https://www.movable-type.co.uk/scripts/latlong.html
func (start GeoPoint) distanceTo(end GeoPoint) float64 {
	earthRadius := 6371e3
	startLatRad := start.Lat * math.Pi / 180
	endLatRad := end.Lat * math.Pi / 180
	deltaLatRad := (end.Lat - start.Lat) * math.Pi / 180
	deltaLongRad := (end.Long - start.Long) * math.Pi / 180

	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(startLatRad)*math.Cos(endLatRad)*
			math.Sin(deltaLongRad/2)*math.Sin(deltaLongRad/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// Alert - data type that
type Alert struct {
	Coord     GeoPoint
	AlertType int
}

func (a Alert) MarshalText() (text []byte, err error) {
	type x Alert
	return json.Marshal(x(a))
}

func (a *Alert) UnmarshalText(text []byte) error {
	type x Alert
	return json.Unmarshal(text, (*x)(a))
}

type AlertData struct {
	Active             bool
	Verified           bool
	Confirmations      float64
	Denies             float64
	Creation           time.Time
	LatestConfirmation time.Time
}

// MarshalText - marshaling for the AlertData struct
func (data AlertData) MarshalText() (text []byte, err error) {
	type x AlertData
	return json.Marshal(x(data))
}

// UnmarshalText - unmarshaling for the AlertData struct
func (data *AlertData) UnmarshalText(text []byte) error {
	type x AlertData
	return json.Unmarshal(text, (*x)(data))
}
