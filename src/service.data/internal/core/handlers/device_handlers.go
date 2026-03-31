package handlers

import (
	"context"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"

	"service.data/internal/components"
	"service.data/internal/core/dtos"
	"service.data/internal/views"
)

const (
	devicesPageSize = 20
)

func (h *Handler) DevicesPageHandler(w http.ResponseWriter, r *http.Request) {
	var (
		stats   []dtos.TileDto
		devices []dtos.DeviceStateDto
	)

	g, ctx := errgroup.WithContext(r.Context())

	g.Go(func() error {
		var err error
		stats, err = h.getDevicesStats(ctx)
		return err
	})

	g.Go(func() error {
		var err error
		devices, err = h.listDevices(ctx, devicesPageSize, "", "")
		return err
	})

	if err := g.Wait(); err != nil {
		h.logger.Error(err, "Error fetching data from ADX: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := views.DevicesView(stats, devices).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Error rendering devices page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) SearchDevicesHandler(w http.ResponseWriter, r *http.Request) {
	if !isHtmx(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	match := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	status := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
	devices, err := h.listDevices(r.Context(), devicesPageSize, match, status)
	if err != nil {
		h.logger.Error(err, "Error fetching devices from ADX: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err = components.DeviceList(devices).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Error rendering device list: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// getDevicesStats retrieves device statistics from ADX.
func (h *Handler) getDevicesStats(ctx context.Context) ([]dtos.TileDto, error) {
	s, err := h.adx.GetDevicesStats(ctx)
	if err != nil {
		h.logger.Error(err, "Error fetching devices stats: %v", err)
		return []dtos.TileDto{}, err
	}

	return dtos.NewTilesFromAdxDevicesStats(s), nil
}

// listDevices retrieves a list of devices from ADX based on the provided page size, match string, and status filter.
func (h *Handler) listDevices(ctx context.Context, pageSize int64, match string, status string) ([]dtos.DeviceStateDto, error) {
	d, err := h.adx.ListDevices(ctx, pageSize, match, status)
	if err != nil {
		h.logger.Error(err, "Error fetching devices from ADX: %v", err)
		return nil, err
	}

	devices := make([]dtos.DeviceStateDto, len(d))
	for i, adxDevice := range d {
		devices[i] = dtos.NewDeviceStateFromAdxDeviceState(adxDevice)
	}

	return devices, nil
}
