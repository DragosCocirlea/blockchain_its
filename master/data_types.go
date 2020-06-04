package main

import (
	"encoding/json"
	"strconv"
	"time"
)

// MessageRegionToMaster - message the region node sends to the master node
type MessageRegionToMaster struct {
	SpeedReports    map[GeoPoint]map[FloatString]float64
	NewAlerts       map[Alert]AlertData
	UsersReputation map[string]float64
}

// FloatString - data type for float that can be marshalled
type FloatString float64

func (fs FloatString) MarshalText() ([]byte, error) {
	vs := strconv.FormatFloat(float64(fs), 'f', 2, 64)
	return []byte(`"` + vs + `"`), nil
}

func (fs *FloatString) UnmarshalText(b []byte) error {
	if b[0] == '"' {
		b = b[1 : len(b)-1]
	}
	f, err := strconv.ParseFloat(string(b), 64)
	*fs = FloatString(f)
	return err
}

// GeoPoint - point in geographic coordinate system
type GeoPoint struct {
	Lat  float64
	Long float64
}

// MarshalText - marshaling for the GeoPoint struct
func (g GeoPoint) MarshalText() (text []byte, err error) {
	type x GeoPoint
	return json.Marshal(x(g))
}

// UnmarshalText - unmarshaling for the GeoPoint struct
func (g *GeoPoint) UnmarshalText(text []byte) error {
	type x GeoPoint
	return json.Unmarshal(text, (*x)(g))
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
