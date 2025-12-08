package models

type Distance struct {
	FromCity string `json:"from_city"`
	ToCity   string `json:"to_city"`
	Km       int    `json:"km"`
}
