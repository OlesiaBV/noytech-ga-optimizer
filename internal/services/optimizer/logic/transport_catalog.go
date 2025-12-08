package logic

import "noytech-ga-optimizer/api/proto"

type TransportSpec struct {
	Type    proto.TransportType
	CapTons float64 // Номинальная грузоподъёмность, тонны
	CapM3   float64 // Номинальный объём кузова, м³
}

var AvailableTransports = []TransportSpec{
	{Type: proto.TransportType_TRANSPORT_1_5T_10M3, CapTons: 1.5, CapM3: 10},
	{Type: proto.TransportType_TRANSPORT_3T_20M3, CapTons: 3.0, CapM3: 20},
	{Type: proto.TransportType_TRANSPORT_5T_36M3, CapTons: 5.0, CapM3: 36},
	{Type: proto.TransportType_TRANSPORT_10T_45M3, CapTons: 10.0, CapM3: 45},
	{Type: proto.TransportType_TRANSPORT_20T_86M3, CapTons: 20.0, CapM3: 86},
}
