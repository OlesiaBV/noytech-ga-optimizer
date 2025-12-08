package models

// Тариф на межгород (руб/км)
type InterCityRate struct {
	VolumeM3   float64 `json:"volume_m3"`
	WeightTons float64 `json:"weight_tons"`
	RatePerKm  float64 `json:"rate_fixed"`
}

// Тариф на внутригород (руб)
type IntraCityRate struct {
	VolumeM3   float64 `json:"volume_m3"`
	WeightTons float64 `json:"weight_tons"`
	RateFixed  float64 `json:"rate_fixed"`
}
