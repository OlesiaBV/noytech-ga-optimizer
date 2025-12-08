package storage

import (
	"context"

	"noytech-ga-optimizer/internal/models"
)

type Storage interface {
	// Shipments
	TruncateShipments(ctx context.Context) error
	BatchInsertShipments(ctx context.Context, shipments []models.Shipment) error

	// Terminals
	TruncateTerminals(ctx context.Context) error
	BatchInsertTerminals(ctx context.Context, terminals []models.Terminal) error

	// Distances
	TruncateDistances(ctx context.Context) error
	BatchInsertDistances(ctx context.Context, distances []models.Distance) error

	// Rates
	TruncateInterCityRates(ctx context.Context) error
	BatchInsertInterCityRates(ctx context.Context, rates []models.InterCityRate) error

	TruncateIntraCityRates(ctx context.Context) error
	BatchInsertIntraCityRates(ctx context.Context, rates []models.IntraCityRate) error

	GetAllShipments(ctx context.Context) ([]models.Shipment, error)
	GetAllTerminals(ctx context.Context) ([]models.Terminal, error)
	GetAllDistances(ctx context.Context) ([]models.Distance, error)
	GetAllInterCityRates(ctx context.Context) ([]models.InterCityRate, error)
	GetAllIntraCityRates(ctx context.Context) ([]models.IntraCityRate, error)
}
