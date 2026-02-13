package handlers

import (
	"net/http"
	"strings"

	"service.data/internal/core/dtos"
	"service.data/internal/views"
)

func (h *Handler) DevicesPageHandler(w http.ResponseWriter, r *http.Request) {
	d, err := h.adx.ListDevices(r.Context(), 20, "")
	if err != nil {
		h.logger.Error(err, "Error fetching devices from ADX: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	devices := make([]dtos.DeviceStateDto, len(d))
	for i, adxDevice := range d {
		devices[i] = dtos.NewFromAdxDevice(adxDevice)
	}

	if err = views.DevicesView(devices).Render(r.Context(), w); err != nil {
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
	d, err := h.adx.ListDevices(r.Context(), 10, match)
	if err != nil {
		h.logger.Error(err, "Error fetching devices from ADX: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	devices := make([]dtos.DeviceStateDto, len(d))
	for i, adxDevice := range d {
		devices[i] = dtos.NewFromAdxDevice(adxDevice)
	}

	if err = views.DeviceList(devices).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Error rendering device list: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
