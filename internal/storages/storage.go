package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"noytech-ga-optimizer/internal/models"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{pool: pool}
}

func (s *PostgresStorage) TruncateShipments(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "TRUNCATE shipments CASCADE")
	return err
}

func (s *PostgresStorage) BatchInsertShipments(ctx context.Context, shipments []models.Shipment) error {
	batch := &pgx.Batch{}
	for _, sh := range shipments {
		batch.Queue(`
			INSERT INTO shipments (id, weight_kg, volume_m3, destination_city, date)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING`,
			sh.ID, sh.WeightKg, sh.VolumeM3, sh.DestinationCity, sh.Date)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) TruncateTerminals(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "TRUNCATE terminals CASCADE")
	return err
}

func (s *PostgresStorage) BatchInsertTerminals(ctx context.Context, terminals []models.Terminal) error {
	batch := &pgx.Batch{}
	for _, t := range terminals {
		batch.Queue(`
			INSERT INTO terminals (city, direction, distance_from_moscow_km)
			VALUES ($1, $2, $3)
			ON CONFLICT (city) DO NOTHING`,
			t.City, t.Direction, t.DistanceFromMoscowKm)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) TruncateDistances(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "TRUNCATE distances CASCADE")
	return err
}

func (s *PostgresStorage) BatchInsertDistances(ctx context.Context, distances []models.Distance) error {
	batch := &pgx.Batch{}
	for _, d := range distances {
		batch.Queue(`
			INSERT INTO distances (from_city, to_city, km)
			VALUES ($1, $2, $3)
			ON CONFLICT (from_city, to_city) DO NOTHING`,
			d.FromCity, d.ToCity, d.Km)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) TruncateInterCityRates(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "TRUNCATE inter_city_rates CASCADE")
	return err
}

func (s *PostgresStorage) BatchInsertInterCityRates(ctx context.Context, rates []models.InterCityRate) error {
	batch := &pgx.Batch{}
	for _, r := range rates {
		batch.Queue(`
			INSERT INTO inter_city_rates (volume_m3, weight_tons, rate_per_km)
			VALUES ($1, $2, $3)
			ON CONFLICT (volume_m3, weight_tons) DO NOTHING`,
			r.VolumeM3, r.WeightTons, r.RatePerKm)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) TruncateIntraCityRates(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "TRUNCATE intra_city_rates CASCADE")
	return err
}

func (s *PostgresStorage) BatchInsertIntraCityRates(ctx context.Context, rates []models.IntraCityRate) error {
	batch := &pgx.Batch{}
	for _, r := range rates {
		batch.Queue(`
			INSERT INTO intra_city_rates (volume_m3, weight_tons, rate_fixed)
			VALUES ($1, $2, $3)
			ON CONFLICT (volume_m3, weight_tons) DO NOTHING`,
			r.VolumeM3, r.WeightTons, r.RateFixed)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) GetAllShipments(ctx context.Context) ([]models.Shipment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, weight_kg, volume_m3, destination_city, date
		FROM shipments
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []models.Shipment
	for rows.Next() {
		var s models.Shipment
		err = rows.Scan(&s.ID, &s.WeightKg, &s.VolumeM3, &s.DestinationCity, &s.Date)
		if err != nil {
			return nil, err
		}
		shipments = append(shipments, s)
	}
	return shipments, nil
}

func (s *PostgresStorage) GetAllTerminals(ctx context.Context) ([]models.Terminal, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT city, direction, distance_from_moscow_km
		FROM terminals
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terminals []models.Terminal
	for rows.Next() {
		var t models.Terminal
		err = rows.Scan(&t.City, &t.Direction, &t.DistanceFromMoscowKm)
		if err != nil {
			return nil, err
		}
		terminals = append(terminals, t)
	}
	return terminals, nil
}

func (s *PostgresStorage) GetAllDistances(ctx context.Context) ([]models.Distance, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT from_city, to_city, km
		FROM distances
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var distances []models.Distance
	for rows.Next() {
		var d models.Distance
		err = rows.Scan(&d.FromCity, &d.ToCity, &d.Km)
		if err != nil {
			return nil, err
		}
		distances = append(distances, d)
	}
	return distances, nil
}

func (s *PostgresStorage) GetAllInterCityRates(ctx context.Context) ([]models.InterCityRate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT volume_m3, weight_tons, rate_per_km
		FROM inter_city_rates
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []models.InterCityRate
	for rows.Next() {
		var r models.InterCityRate
		err = rows.Scan(&r.VolumeM3, &r.WeightTons, &r.RatePerKm)
		if err != nil {
			return nil, err
		}
		rates = append(rates, r)
	}
	return rates, nil
}

func (s *PostgresStorage) GetAllIntraCityRates(ctx context.Context) ([]models.IntraCityRate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT volume_m3, weight_tons, rate_fixed
		FROM intra_city_rates
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []models.IntraCityRate
	for rows.Next() {
		var r models.IntraCityRate
		err = rows.Scan(&r.VolumeM3, &r.WeightTons, &r.RateFixed)
		if err != nil {
			return nil, err
		}
		rates = append(rates, r)
	}
	return rates, nil
}
