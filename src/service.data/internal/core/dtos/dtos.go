package dtos

import (
	"fmt"

	"service.data/external/storage/adx"
)

// TileDto represents a single tile.
type TileDto struct {
	Title string
	Value string
	Desc  string
	Icon  string
	Type  string
}

// NewTilesFromAdxDevicesStats creates a list of TileDto instances from an adx.DevicesStats.
func NewTilesFromAdxDevicesStats(d adx.DevicesStats) []TileDto {
	return []TileDto{
		{
			Title: "Total Devices",
			Value: fmt.Sprintf("%d", d.Total),
			Desc:  "All registered devices",
			Icon:  "device",
			Type:  "info",
		},
		{
			Title: "Online",
			Value: fmt.Sprintf("%d", d.Online),
			Desc:  fmt.Sprintf("%d%% active", d.Online*100/d.Total),
			Icon:  "online",
			Type:  "success",
		},
		{
			Title: "Offline",
			Value: fmt.Sprintf("%d", d.Offline),
			Desc:  fmt.Sprintf("%d%% inactive", d.Offline*100/d.Total),
			Icon:  "offline",
			Type:  "error",
		},
		{
			Title: "Issues",
			Value: fmt.Sprintf("%d", d.Issues),
			Desc:  "Require attention",
			Icon:  "warning",
			Type:  "warning",
		},
		{
			Title: "Total Sensors",
			Value: fmt.Sprintf("%d", d.TotalSensors),
			Desc:  "Across all devices",
			Icon:  "sensor",
			Type:  "default",
		},
	}
}
