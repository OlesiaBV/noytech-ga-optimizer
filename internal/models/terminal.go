package models

type Terminal struct {
	City                 string `json:"city"`
	Direction            string `json:"direction"`
	DistanceFromMoscowKm int    `json:"distance_from_moscow_km"`
}
