package validation

import (
	"fmt"
	"strings"

	"noytech-ga-optimizer/api/proto"
	"noytech-ga-optimizer/pkg/errors"
)

var AllowedDays = map[string]bool{
	"mon": true, "tue": true, "wed": true, "thu": true,
	"fri": true, "sat": true, "sun": true,
}

var AllowedDirections = map[string]bool{
	"Восток":       true,
	"Северо-Запад": true,
	"Юг":           true,
	"Волга":        true,
}

var AllowedSelectionTypes = map[proto.SelectionType]bool{
	proto.SelectionType_SELECTION_TOURNAMENT: true,
	proto.SelectionType_SELECTION_ROULETTE:   true,
	proto.SelectionType_SELECTION_RANK:       true,
}

var AllowedCrossoverTypes = map[proto.CrossoverType]bool{
	proto.CrossoverType_CROSSOVER_UNIFORM:      true,
	proto.CrossoverType_CROSSOVER_SINGLE_POINT: true,
	proto.CrossoverType_CROSSOVER_TWO_POINT:    true,
}

var AllowedMutationTypes = map[proto.MutationType]bool{
	proto.MutationType_MUTATION_INVERSION: true,
	proto.MutationType_MUTATION_SWAP:      true,
}

func ValidateOptimizeRequest(req *proto.OptimizeRequest) error {
	if req == nil {
		return errors.NewErrInvalidArgument(nil, "request body is required")
	}

	var validationErrors []errors.ErrorDetail

	// 1. direction (необязательное, но если указано — должно быть в списке)
	if req.Direction != "" {
		if !AllowedDirections[req.Direction] {
			allowed := strings.Join(allowedKeys(AllowedDirections), ", ")
			validationErrors = append(validationErrors, errors.ErrorDetail{
				Field:   "direction",
				Message: fmt.Sprintf("value '%s' is not allowed. Allowed: %s", req.Direction, allowed),
			})
		}
	}

	// 2. delivery_days (обязательно: ровно 2 уникальных дня из списка)
	if len(req.DeliveryDays) == 0 {
		validationErrors = append(validationErrors, errors.ErrorDetail{
			Field:   "delivery_days",
			Message: "field is required",
		})
	} else {
		if len(req.DeliveryDays) != 2 {
			allowed := strings.Join(allowedKeys(AllowedDays), ", ")
			validationErrors = append(validationErrors, errors.ErrorDetail{
				Field:   "delivery_days",
				Message: fmt.Sprintf("exactly 2 delivery days must be provided, got %d. Allowed values: %s", len(req.DeliveryDays), allowed),
			})
		} else {
			seen := make(map[string]bool)
			for i, day := range req.DeliveryDays {
				trimmed := strings.TrimSpace(day)
				if trimmed == "" {
					validationErrors = append(validationErrors, errors.ErrorDetail{
						Field:   fmt.Sprintf("delivery_days[%d]", i),
						Message: "day cannot be empty",
					})
					continue
				}
				lower := strings.ToLower(trimmed)
				if !AllowedDays[lower] {
					allowed := strings.Join(allowedKeys(AllowedDays), ", ")
					validationErrors = append(validationErrors, errors.ErrorDetail{
						Field:   fmt.Sprintf("delivery_days[%d]", i),
						Message: fmt.Sprintf("invalid day '%s'. Allowed: %s", day, allowed),
					})
				} else if seen[lower] {
					validationErrors = append(validationErrors, errors.ErrorDetail{
						Field:   fmt.Sprintf("delivery_days[%d]", i),
						Message: fmt.Sprintf("duplicate delivery day: '%s'", day),
					})
				}
				seen[lower] = true
			}
		}
	}

	// 3. ga_settings_level_1
	if req.GaSettingsLevel_1 == nil {
		validationErrors = append(validationErrors, errors.ErrorDetail{
			Field:   "ga_settings_level_1",
			Message: "field is required",
		})
	} else {
		gaErrs := validateGASettings(req.GaSettingsLevel_1, "ga_settings_level_1")
		validationErrors = append(validationErrors, gaErrs...)
	}

	if len(validationErrors) > 0 {
		return errors.NewErrInvalidArgumentWithDetails(validationErrors)
	}

	return nil
}

func validateGASettings(settings *proto.GASettings, prefix string) []errors.ErrorDetail {
	var errs []errors.ErrorDetail

	// num_generations
	if settings.NumGenerations <= 0 {
		errs = append(errs, errors.ErrorDetail{
			Field:   prefix + ".num_generations",
			Message: "must be greater than 0",
		})
	}

	// num_individuals
	if settings.NumIndividuals <= 0 {
		errs = append(errs, errors.ErrorDetail{
			Field:   prefix + ".num_individuals",
			Message: "must be greater than 0",
		})
	}

	// stopping_criterion
	if settings.StoppingCriterion <= 0 {
		errs = append(errs, errors.ErrorDetail{
			Field:   prefix + ".stopping_criterion",
			Message: "must be greater than 0",
		})
	}

	// selection_type
	if settings.SelectionType == proto.SelectionType_SELECTION_UNSPECIFIED {
		allowed := strings.Join(allowedEnumValuesSelection(), ", ")
		errs = append(errs, errors.ErrorDetail{
			Field:   prefix + ".selection_type",
			Message: fmt.Sprintf("field is required. Allowed values: %s", allowed),
		})
	} else if !AllowedSelectionTypes[settings.SelectionType] {
		allowed := strings.Join(allowedEnumValuesSelection(), ", ")
		errs = append(errs, errors.ErrorDetail{
			Field:   prefix + ".selection_type",
			Message: fmt.Sprintf("invalid value. Allowed: %s", allowed),
		})
	}

	// crossover_type
	if settings.CrossoverType == proto.CrossoverType_CROSSOVER_UNSPECIFIED {
		allowed := strings.Join(allowedEnumValuesCrossover(), ", ")
		errs = append(errs, errors.ErrorDetail{
			Field:   prefix + ".crossover_type",
			Message: fmt.Sprintf("field is required. Allowed values: %s", allowed),
		})
	} else if !AllowedCrossoverTypes[settings.CrossoverType] {
		allowed := strings.Join(allowedEnumValuesCrossover(), ", ")
		errs = append(errs, errors.ErrorDetail{
			Field:   prefix + ".crossover_type",
			Message: fmt.Sprintf("invalid value. Allowed: %s", allowed),
		})
	}

	// mutation_type
	if settings.MutationType == proto.MutationType_MUTATION_UNSPECIFIED {
		allowed := strings.Join(allowedEnumValuesMutation(), ", ")
		errs = append(errs, errors.ErrorDetail{
			Field:   prefix + ".mutation_type",
			Message: fmt.Sprintf("field is required. Allowed values: %s", allowed),
		})
	} else if !AllowedMutationTypes[settings.MutationType] {
		allowed := strings.Join(allowedEnumValuesMutation(), ", ")
		errs = append(errs, errors.ErrorDetail{
			Field:   prefix + ".mutation_type",
			Message: fmt.Sprintf("invalid value. Allowed: %s", allowed),
		})
	}

	return errs
}

func allowedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func allowedEnumValuesSelection() []string {
	keys := make([]string, 0, len(AllowedSelectionTypes))
	for k := range AllowedSelectionTypes {
		keys = append(keys, k.String())
	}
	return keys
}

func allowedEnumValuesCrossover() []string {
	keys := make([]string, 0, len(AllowedCrossoverTypes))
	for k := range AllowedCrossoverTypes {
		keys = append(keys, k.String())
	}
	return keys
}

func allowedEnumValuesMutation() []string {
	keys := make([]string, 0, len(AllowedMutationTypes))
	for k := range AllowedMutationTypes {
		keys = append(keys, k.String())
	}
	return keys
}
