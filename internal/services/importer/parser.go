package importer

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/xuri/excelize/v2"

	"noytech-ga-optimizer/internal/models"
	storage "noytech-ga-optimizer/internal/storages"
	"noytech-ga-optimizer/pkg/errors"
)

type SheetConfig struct {
	Name       string
	Directions []string
}

var sheetConfigs = []SheetConfig{
	{Name: "Северо-Запад", Directions: []string{"М10"}},
	{Name: "Восток", Directions: []string{"М4", "М7"}},
	{Name: "Волга", Directions: []string{"М6"}},
	{Name: "Юг", Directions: []string{"М5"}},
}

func parseAndLoadShipmentsAndTerminals(ctx context.Context, storage storage.Storage, f *excelize.File, logger *slog.Logger) error {
	logger = logger.With(slog.String("submethod", "parseAndLoadShipmentsAndTerminals"))

	// Загрузка shipments из листа Data
	rows, err := f.GetRows("Data")
	if err != nil {
		logger.Warn("Sheet 'Data' not found, skipping shipments", "error", err)
		return nil
	}

	var shipments []models.Shipment
	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 8 {
			continue
		}

		weight, errW := strconv.ParseFloat(row[1], 64)
		volume, errV := strconv.ParseFloat(row[2], 64)

		if errW != nil || errV != nil {
			logger.Warn("Skipping shipment due to invalid weight or volume", "row_index", i, "row", row)
			continue
		}

		if weight <= 0 {
			logger.Warn("Skipping shipment due to non-positive weight_kg", "row_index", i, "weight_kg", weight)
			continue
		}
		if volume <= 0 {
			logger.Warn("Skipping shipment due to non-positive volume_m3", "row_index", i, "volume_m3", volume)
			continue
		}

		dateStr := strings.TrimSpace(row[7])
		date, err := parseDate(dateStr)
		if err != nil {
			logger.Warn("Could not parse date, skipping row", "row_index", i, "date", dateStr, "error", err)
			continue
		}

		shipment := models.Shipment{
			ID:              strings.TrimSpace(row[0]),
			WeightKg:        weight,
			VolumeM3:        volume,
			DestinationCity: strings.TrimSpace(row[3]),
			Date:            date,
		}

		shipments = append(shipments, shipment)
	}

	if len(shipments) > 0 {
		if err := storage.BatchInsertShipments(ctx, shipments); err != nil {
			if pqErr, ok := err.(*pgconn.PgError); ok {
				if pqErr.Code == "23514" {
					if strings.Contains(pqErr.Detail, "shipments_volume_m3_check") {
						return errors.NewUnprocessableEntityError("Ошибка при загрузке файла 'stat.xlsx': в листе 'Data' обнаружены строки с невалидным значением 'М3' (должно быть > 0).")
					}
					if strings.Contains(pqErr.Detail, "shipments_weight_kg_check") {
						return errors.NewUnprocessableEntityError("Ошибка при загрузке файла 'stat.xlsx': в листе 'Data' обнаружены строки с невалидным значением 'Расчетный вес, кг' (должно быть > 0).")
					}
				}
			}
			return errors.NewErrImportShipmentsFailed(err)
		}
		logger.Info("Inserted shipments", "count", len(shipments))
	} else {
		logger.Warn("No valid shipments found in 'Data' sheet")
	}

	// Загрузка terminals из листа Zones
	rows, err = f.GetRows("Zones")
	if err != nil {
		logger.Warn("Sheet 'Zones' not found, skipping terminals", "error", err)
		return nil
	}

	var terminals []models.Terminal
	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 5 {
			continue
		}

		city := strings.TrimSpace(row[1])
		if city == "" {
			logger.Warn("Skipping terminal due to empty city", "row_index", i, "row", row)
			continue
		}

		direction := strings.TrimSpace(row[2])
		if direction == "" {
			logger.Warn("Skipping terminal due to empty direction", "row_index", i, "city", city)
			continue
		}

		distanceStr := strings.TrimSpace(row[3])
		distance, err := strconv.Atoi(distanceStr)
		if err != nil {
			logger.Warn("Skipping terminal due to invalid distance format", "row_index", i, "city", city, "value", distanceStr, "error", err)
			continue
		}
		if distance < 0 {
			logger.Warn("Skipping terminal due to negative distance", "row_index", i, "city", city, "distance", distance)
			continue
		}

		terminal := models.Terminal{
			City:                 city,
			Direction:            direction,
			DistanceFromMoscowKm: distance,
		}

		terminals = append(terminals, terminal)
	}

	if len(terminals) > 0 {
		if err := storage.BatchInsertTerminals(ctx, terminals); err != nil {
			return errors.NewErrImportTerminalsFailed(err)
		}
		logger.Info("Inserted terminals", "count", len(terminals))
	} else {
		logger.Warn("No valid terminals found in 'Zones' sheet")
	}

	return nil
}

func parseAndLoadInterCityRates(ctx context.Context, storage storage.Storage, f *excelize.File, logger *slog.Logger) error {
	logger = logger.With(slog.String("submethod", "parseAndLoadInterCityRates"))

	rows, err := f.GetRows("Тариф на межгород")
	if err != nil {
		logger.Warn("Sheet 'Тариф на межгород' not found, skipping", "error", err)
		return nil
	}

	if len(rows) < 3 {
		logger.Warn("Sheet 'Тариф на межгород' has less than 3 rows, skipping")
		return nil
	}

	volumes := make([]float64, 0)
	weights := make([]float64, 0)
	rates := make([]float64, 0)

	for j := 1; j < len(rows[1]); j++ {
		v, err := strconv.ParseFloat(strings.Replace(rows[1][j], ",", ".", -1), 64)
		if err != nil {
			continue
		}
		volumes = append(volumes, v)
	}

	for j := 1; j < len(rows[2]); j++ {
		w, err := strconv.ParseFloat(strings.Replace(rows[2][j], ",", ".", -1), 64)
		if err != nil {
			continue
		}
		weights = append(weights, w)
	}

	for j := 1; j < len(rows[3]); j++ {
		r, err := strconv.ParseFloat(strings.Replace(rows[3][j], ",", ".", -1), 64)
		if err != nil {
			continue
		}
		rates = append(rates, r)
	}

	if len(volumes) != len(weights) || len(weights) != len(rates) {
		logger.Warn("Mismatched lengths in inter-city rates data")
		return errors.NewUnprocessableEntityError("Ошибка при загрузке 'Тариф на межгород': количество значений в строках 'Объем', 'Масса' и 'Тариф' не совпадает.")
	}

	var rateList []models.InterCityRate
	for i := 0; i < len(volumes); i++ {
		rateList = append(rateList, models.InterCityRate{
			VolumeM3:   volumes[i],
			WeightTons: weights[i],
			RatePerKm:  rates[i],
		})
	}

	if len(rateList) > 0 {
		if err := storage.BatchInsertInterCityRates(ctx, rateList); err != nil {
			if pqErr, ok := err.(*pgconn.PgError); ok {
				if pqErr.Code == "23514" {
					if strings.Contains(pqErr.Detail, "inter_city_rates_rate_per_km_check") {
						return errors.NewUnprocessableEntityError("Ошибка при загрузке 'Тариф на межгород': обнаружены строки с невалидным значением 'руб/км'.")
					}
				}
			}
			return errors.NewErrImportRatesFailed(err)
		}
		logger.Info("Inserted inter-city rates", "count", len(rateList))
	}

	return nil
}

func parseAndLoadIntraCityRates(ctx context.Context, storage storage.Storage, f *excelize.File, logger *slog.Logger) error {
	logger = logger.With(slog.String("submethod", "parseAndLoadIntraCityRates"))

	rows, err := f.GetRows("Тариф на внутригород")
	if err != nil {
		logger.Warn("Sheet 'Тариф на внутригород' not found, skipping", "error", err)
		return nil
	}

	if len(rows) < 3 {
		logger.Warn("Sheet 'Тариф на внутригород' has less than 3 rows, skipping")
		return nil
	}

	volumes := make([]float64, 0)
	weights := make([]float64, 0)
	rates := make([]float64, 0)

	for j := 1; j < len(rows[1]); j++ {
		v, err := strconv.ParseFloat(strings.Replace(rows[1][j], ",", ".", -1), 64)
		if err != nil {
			continue
		}
		volumes = append(volumes, v)
	}

	for j := 1; j < len(rows[2]); j++ {
		w, err := strconv.ParseFloat(strings.Replace(rows[2][j], ",", ".", -1), 64)
		if err != nil {
			continue
		}
		weights = append(weights, w)
	}

	for j := 1; j < len(rows[3]); j++ {
		r, err := strconv.ParseFloat(strings.Replace(rows[3][j], ",", ".", -1), 64)
		if err != nil {
			continue
		}
		rates = append(rates, r)
	}

	if len(volumes) != len(weights) || len(weights) != len(rates) {
		logger.Warn("Mismatched lengths in intra-city rates data")
		return errors.NewUnprocessableEntityError("Ошибка при загрузке 'Тариф на внутригород': количество значений в строках 'Объем', 'Масса' и 'Тариф' не совпадает.")
	}

	var rateList []models.IntraCityRate
	for i := 0; i < len(volumes); i++ {
		rateList = append(rateList, models.IntraCityRate{
			VolumeM3:   volumes[i],
			WeightTons: weights[i],
			RateFixed:  rates[i],
		})
	}

	if len(rateList) > 0 {
		if err := storage.BatchInsertIntraCityRates(ctx, rateList); err != nil {
			if pqErr, ok := err.(*pgconn.PgError); ok {
				if pqErr.Code == "23514" {
					if strings.Contains(pqErr.Detail, "intra_city_rates_rate_fixed_check") {
						return errors.NewUnprocessableEntityError("Ошибка при загрузке 'Тариф на внутригород': обнаружены строки с невалидным значением 'руб'.")
					}
				}
			}
			return errors.NewErrImportRatesFailed(err)
		}
		logger.Info("Inserted intra-city rates", "count", len(rateList))
	}

	return nil
}

func parseAndLoadDistances(ctx context.Context, storage storage.Storage, f *excelize.File, logger *slog.Logger) error {
	logger = logger.With(slog.String("submethod", "parseAndLoadDistances"))

	var allDistances []models.Distance

	for _, config := range sheetConfigs {
		sheetName := config.Name
		if !sheetExists(f, sheetName) {
			logger.Warn("Sheet not found in file, skipping", "sheet", sheetName)
			continue
		}

		rows, err := f.GetRows(sheetName)
		if err != nil {
			logger.Warn("Could not read rows from sheet, skipping", "sheet", sheetName, "error", err)
			continue
		}

		if len(rows) == 0 {
			logger.Warn("Sheet is empty, skipping", "sheet", sheetName)
			continue
		}

		headers := rows[0]
		for i, row := range rows {
			if i == 0 {
				continue
			}

			fromCity := strings.TrimSpace(headers[i])
			if fromCity == "" {
				logger.Warn("Skipping cell: fromCity is empty", "sheet", sheetName, "row_index", i, "col_index", "header")
				continue
			}

			for j, cell := range row {
				if j == 0 {
					continue
				}

				toCity := strings.TrimSpace(headers[j])
				if toCity == "" {
					logger.Warn("Skipping cell: toCity is empty", "sheet", sheetName, "from_city", fromCity, "row_index", i, "col_index", j)
					continue
				}

				cellValue := strings.TrimSpace(cell)

				km, err := strconv.Atoi(cellValue)
				if err != nil {
					logger.Warn("Skipping cell: could not parse distance as integer", "sheet", sheetName, "from_city", fromCity, "to_city", toCity, "value", cellValue, "error", err)
					continue
				}

				allDistances = append(allDistances, models.Distance{
					FromCity: fromCity,
					ToCity:   toCity,
					Km:       km,
				})
			}
		}
	}

	if len(allDistances) > 0 {
		if err := storage.BatchInsertDistances(ctx, allDistances); err != nil {
			if pqErr, ok := err.(*pgconn.PgError); ok {
				if pqErr.Code == "23514" {
					if strings.Contains(pqErr.Detail, "distances_km_check") {
						return errors.NewUnprocessableEntityError("Ошибка при загрузке 'filled_distances_MKR.xlsx': обнаружены строки с отрицательным расстоянием.")
					}
				}
			}
			return errors.NewErrImportDistancesFailed(err)
		}
		logger.Info("Inserted distances", "count", len(allDistances))
	} else {
		logger.Warn("No distances were parsed from the file")
	}

	return nil
}

var supportedDateFormats = []string{
	"2006-01-02",
	"2-Jan-06",
	"1/2/06",
	"01-02-06",
	"02.01.2006",
	"02.01.06",
	"2. January 2006",
}

var russianMonthMap = map[string]string{
	"янв": "01", "фев": "02", "мар": "03", "апр": "04", "май": "05", "июн": "06",
	"июл": "07", "авг": "08", "сен": "09", "окт": "10", "ноя": "11", "дек": "12",
}

func parseDate(dateStr string) (time.Time, error) {
	trimmedDateStr := strings.TrimSpace(dateStr)

	if trimmedDateStr == "" {
		return time.Time{}, fmt.Errorf("date string is empty")
	}

	for _, format := range supportedDateFormats {
		if t, err := time.Parse(format, trimmedDateStr); err == nil {
			return t, nil
		}
	}

	parts := strings.Split(trimmedDateStr, ".")
	if len(parts) == 3 {
		day := parts[0]
		month := strings.ToLower(strings.TrimSpace(parts[1]))
		year := parts[2]

		numericMonth, ok := russianMonthMap[month]
		if !ok {
			return time.Time{}, fmt.Errorf("unknown russian month: %s", month)
		}

		var goDateString string
		var goLayout string

		if len(year) == 2 {
			goDateString = fmt.Sprintf("%s.%s.%s", day, numericMonth, year)
			goLayout = "02.01.06"
		} else if len(year) == 4 {
			goDateString = fmt.Sprintf("%s.%s.%s", day, numericMonth, year)
			goLayout = "02.01.2006"
		} else {
			return time.Time{}, fmt.Errorf("invalid year format in russian date: %s", year)
		}

		return time.Parse(goLayout, goDateString)
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
