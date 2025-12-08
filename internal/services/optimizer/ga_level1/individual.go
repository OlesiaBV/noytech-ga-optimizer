package ga_level1

import "noytech-ga-optimizer/api/proto"

type Individual struct {
	TerminalMask    []bool
	Fitness         float64
	Cost            CostBreakdown
	ActiveTerminals []string
	Routes          []RouteWithShipments
}

type RouteWithShipments struct {
	FromCity      string
	ToTerminal    string
	ShipmentIDs   []string
	Cost          float64
	TransportUsed proto.TransportType
}

type CostBreakdown struct {
	LinehaulCost float64
	LastMileCost float64
	PenaltyCost  float64
	TotalCost    float64
}
