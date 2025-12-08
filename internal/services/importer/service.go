package importer

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/xuri/excelize/v2"

	storage "noytech-ga-optimizer/internal/storages"
	"noytech-ga-optimizer/pkg/errors"
)

type FileData struct {
	Name    string
	Content []byte
}

type Service struct {
	storage storage.Storage
	logger  *slog.Logger
}

func New(s storage.Storage, l *slog.Logger) *Service {
	return &Service{
		storage: s,
		logger:  l,
	}
}

func (svc *Service) ImportFromXLSX(ctx context.Context, files []FileData) error {
	logger := svc.logger.With(slog.String("method", "ImportFromXLSX"))
	logger.Info("Starting data import from XLSX files", "file_count", len(files))

	// 1. Очистка БД
	logger.Info("Truncating all related tables")
	if err := svc.truncateAll(ctx); err != nil {
		logger.Error("Failed to truncate tables", "error", err)
		return errors.NewErrInternal(err, "failed to truncate tables")
	}

	var statFile *FileData
	var distancesFile *FileData

	for _, file := range files {
		f, err := excelize.OpenReader(bytes.NewReader(file.Content))
		if err != nil {
			logger.Warn("Could not open file as XLSX, skipping", "filename", file.Name, "error", err)
			continue
		}

		if sheetExists(f, "Data") || sheetExists(f, "Zones") {
			statFile = &file
		} else if sheetExists(f, "Восток") || sheetExists(f, "Волга") || sheetExists(f, "Юг") || sheetExists(f, "Северо-Запад") {
			distancesFile = &file
		}
		f.Close()
	}

	if statFile != nil {
		logger.Info("Processing stat file", "filename", statFile.Name)
		if err := svc.processStatFile(ctx, statFile); err != nil {
			logger.Error("Failed to process stat file", "filename", statFile.Name, "error", err)
			return fmt.Errorf("process stat file: %w", err)
		} else {
			logger.Info("Successfully processed stat file", "filename", statFile.Name)
		}
	}

	if distancesFile != nil {
		logger.Info("Processing distances file", "filename", distancesFile.Name)
		if err := svc.processDistancesFile(ctx, distancesFile); err != nil {
			logger.Error("Failed to process distances file", "filename", distancesFile.Name, "error", err)
			return fmt.Errorf("process distances file: %w", err)
		} else {
			logger.Info("Successfully processed distances file", "filename", distancesFile.Name)
		}
	}

	logger.Info("Data import completed")
	return nil
}

func (svc *Service) truncateAll(ctx context.Context) error {
	if err := svc.storage.TruncateShipments(ctx); err != nil {
		return fmt.Errorf("truncate shipments: %w", err)
	}
	if err := svc.storage.TruncateTerminals(ctx); err != nil {
		return fmt.Errorf("truncate terminals: %w", err)
	}
	if err := svc.storage.TruncateDistances(ctx); err != nil {
		return fmt.Errorf("truncate distances: %w", err)
	}
	if err := svc.storage.TruncateInterCityRates(ctx); err != nil {
		return fmt.Errorf("truncate inter_city_rates: %w", err)
	}
	if err := svc.storage.TruncateIntraCityRates(ctx); err != nil {
		return fmt.Errorf("truncate intra_city_rates: %w", err)
	}
	return nil
}

func (svc *Service) processStatFile(ctx context.Context, file *FileData) error {
	logger := svc.logger.With(slog.String("method", "processStatFile"), slog.String("filename", file.Name))

	f, err := excelize.OpenReader(bytes.NewReader(file.Content))
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	if err := svc.loadShipmentsAndTerminals(ctx, f, logger); err != nil {
		return fmt.Errorf("load shipments/terminals: %w", err)
	}

	if err := svc.loadInterCityRates(ctx, f, logger); err != nil {
		return fmt.Errorf("load inter-city rates: %w", err)
	}

	if err := svc.loadIntraCityRates(ctx, f, logger); err != nil {
		return fmt.Errorf("load intra-city rates: %w", err)
	}

	return nil
}

func (svc *Service) processDistancesFile(ctx context.Context, file *FileData) error {
	logger := svc.logger.With(slog.String("method", "processDistancesFile"), slog.String("filename", file.Name))

	f, err := excelize.OpenReader(bytes.NewReader(file.Content))
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	if err := svc.loadDistances(ctx, f, logger); err != nil {
		return fmt.Errorf("load distances: %w", err)
	}

	return nil
}

func (svc *Service) loadShipmentsAndTerminals(ctx context.Context, f *excelize.File, logger *slog.Logger) error {
	return parseAndLoadShipmentsAndTerminals(ctx, svc.storage, f, logger)
}

func (svc *Service) loadInterCityRates(ctx context.Context, f *excelize.File, logger *slog.Logger) error {
	return parseAndLoadInterCityRates(ctx, svc.storage, f, logger)
}

func (svc *Service) loadIntraCityRates(ctx context.Context, f *excelize.File, logger *slog.Logger) error {
	return parseAndLoadIntraCityRates(ctx, svc.storage, f, logger)
}

func (svc *Service) loadDistances(ctx context.Context, f *excelize.File, logger *slog.Logger) error {
	return parseAndLoadDistances(ctx, svc.storage, f, logger)
}

func sheetExists(f *excelize.File, sheetName string) bool {
	sheets := f.GetSheetMap()
	for _, name := range sheets {
		if name == sheetName {
			return true
		}
	}
	return false
}
