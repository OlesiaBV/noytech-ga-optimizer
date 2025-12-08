package logic

import (
	"fmt"
	"time"

	"noytech-ga-optimizer/internal/models"
)

var DeliveryDayMap = map[string]time.Weekday{
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
	"sun": time.Sunday,
}

func GroupShipmentsByDeliveryDay(shipments []models.Shipment, deliveryDays []string) (map[string][]models.Shipment, error) {
	if len(deliveryDays) != 2 {
		return nil, fmt.Errorf("exactly 2 delivery days are required, got %d", len(deliveryDays))
	}

	day1Str, day2Str := deliveryDays[0], deliveryDays[1]
	day1, ok1 := DeliveryDayMap[day1Str]
	day2, ok2 := DeliveryDayMap[day2Str]
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("invalid delivery day(s): %v. Allowed: mon, tue, wed, thu, fri, sat, sun", deliveryDays)
	}

	result := map[string][]models.Shipment{
		day1Str: make([]models.Shipment, 0),
		day2Str: make([]models.Shipment, 0),
	}

	var interval1Start, interval1End, interval2Start, interval2End time.Weekday
	var key1, key2 string

	if day1 < day2 {
		interval1Start = time.Saturday
		interval1End = day1
		interval2Start = day1 + 1
		interval2End = day2
		key1 = day1Str
		key2 = day2Str
	} else {
		interval1Start = day2 + 1
		interval1End = day1
		interval2Start = time.Saturday
		interval2End = day2
		key1 = day1Str
		key2 = day2Str
	}

	for _, shipment := range shipments {
		shipDay := shipment.Date.Weekday()

		if isBetween(shipDay, interval1Start, interval1End) {
			result[key1] = append(result[key1], shipment)
		} else if isBetween(shipDay, interval2Start, interval2End) {
			result[key2] = append(result[key2], shipment)
		}
	}

	return result, nil
}

func isBetween(day, start, end time.Weekday) bool {
	if start <= end {
		return day >= start && day <= end
	}
	return day >= start || day <= end
}
