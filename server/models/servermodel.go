package models

const MAX_TIME_TO_STAY_ALIVE = 30


// ServerModel is a model of a file
type ServerModel struct {
	ID          string         `json:"id"`
	IP          string         `json:"ip"`
	IsAlive    bool        `json:"isAlive"`
	Latency     float64      `json:"latency"`
	TimeSinceLastAlive        float64      `json:"timeSinceLastAlive"`
}

