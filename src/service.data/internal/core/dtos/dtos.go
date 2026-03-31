package dtos

import (
	"fmt"

	"service.data/external/storage/adx"
)

// TileDto represents a single tile.
type TileDto struct {
	Title     string
	Value     string
	Desc      string
	Icon      string
	Type      string
	FilterKey string
}

// NewTilesFromAdxDevicesStats creates a list of TileDto instances from an adx.DevicesStats.
func NewTilesFromAdxDevicesStats(d adx.DevicesStats) []TileDto {
	pctActive := int64(0)
	pctInactive := int64(0)
	if d.Total > 0 {
		pctActive = d.Online * 100 / d.Total
		pctInactive = d.Offline * 100 / d.Total
	}

	return []TileDto{
		{
			Title:     "Total Devices",
			Value:     fmt.Sprintf("%d", d.Total),
			Desc:      "All registered devices",
			Icon:      "device",
			Type:      "info",
			FilterKey: "",
		},
		{
			Title:     "Online",
			Value:     fmt.Sprintf("%d", d.Online),
			Desc:      fmt.Sprintf("%d%% active", pctActive),
			Icon:      "online",
			Type:      "success",
			FilterKey: "on",
		},
		{
			Title:     "Offline",
			Value:     fmt.Sprintf("%d", d.Offline),
			Desc:      fmt.Sprintf("%d%% inactive", pctInactive),
			Icon:      "offline",
			Type:      "error",
			FilterKey: "off",
		},
		{
			Title:     "Issues",
			Value:     fmt.Sprintf("%d", d.Issues),
			Desc:      "Require attention",
			Icon:      "warning",
			Type:      "warning",
			FilterKey: "issue",
		},
		{
			Title:     "Total Sensors",
			Value:     fmt.Sprintf("%d", d.TotalSensors),
			Desc:      "Across all devices",
			Icon:      "sensor",
			Type:      "default",
			FilterKey: "",
		},
	}
}
