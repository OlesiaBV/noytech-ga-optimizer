package models

import "time"

type Shipment struct {
	ID              string    `json:"id"`
	WeightKg        float64   `json:"weight_kg"`
	VolumeM3        float64   `json:"volume_m3"`
	DestinationCity string    `json:"destination_city"`
	Date            time.Time `json:"date"`
}
