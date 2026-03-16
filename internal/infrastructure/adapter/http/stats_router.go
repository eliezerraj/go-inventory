package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/go-inventory/internal/domain/model"
)

// About get inventory
func (h *HttpRouters) GetInventoryTimeSeries(rw http.ResponseWriter, req *http.Request) error {			
	ctx, cancel, span := h.withContext(req, "GetInventoryTimeSeries")
	defer cancel()
	defer span.End()

	// decode payload	
	vars := mux.Vars(req)
	varID := vars["id"]
	inventory := model.Inventory{Product: model.Product{Sku: varID}}

	// call service	
	res, err := h.workerService.GetInventoryTimeSeries(ctx, 72, &inventory)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusOK, res)
}