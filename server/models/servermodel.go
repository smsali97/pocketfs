package models

import "time"

const MAX_TIME_TO_STAY_ALIVE = 30 // in seconds


// ServerModel is a model of a file
type ServerModel struct {
	ID          string         `json:"id"`
	IP          string         `json:"ip"`
	Port		string				`json:"port"`
	IsAlive    bool        `json:"isAlive"`
	Latency     float64      `json:"latency"`
	NoOfPings   int			`json:"noOfPings""`
	TimeSinceLastAlive        float64      `json:"timeSinceLastAlive"`
	LastSeen	time.Time	`json:"lastSeen"`
}

