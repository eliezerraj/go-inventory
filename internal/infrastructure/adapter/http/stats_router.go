package http

import (
	"net/http"
	"strconv"

	"github.com/go-inventory/internal/domain/model"
	"github.com/go-inventory/shared/erro"
)

// About get timeseries inventory data for a product
func (h *HttpRouters) GetInventoryTimeSeries(rw http.ResponseWriter, req *http.Request) error {			
	ctx, cancel, span := h.withContext(req, "GetInventoryTimeSeries")
	defer cancel()
	defer span.End()

	query := req.URL.Query()
	sku := query.Get("sku")
	if sku == "" {
		return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
	}

	// default window is 24, can be override by query parameter
	window := 24
	windowParam := query.Get("window")
	if windowParam != "" {
		parsedWindow, err := strconv.Atoi(windowParam)
		if err != nil || parsedWindow <= 0 {
			return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
		}
		window = parsedWindow
	}

	inventory := model.Inventory{Product: model.Product{Sku: sku}}

	// call service	
	res, err := h.workerService.GetInventoryTimeSeries(ctx, window, &inventory)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusOK, res)
}