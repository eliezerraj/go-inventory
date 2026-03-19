package http

import (
	"net/http"
	"strconv"
	"encoding/json"	

	"go.opentelemetry.io/otel/codes"	
	"github.com/gorilla/mux"
	"github.com/go-inventory/internal/domain/model"
	"github.com/go-inventory/shared/erro"
)

// About get inventory
func (h *HttpRouters) GetInventory(rw http.ResponseWriter, req *http.Request) error {			
	ctx, cancel, span := h.withContext(req, "GetInventory")
	defer cancel()
	defer span.End()

	// decode payload	
	vars := mux.Vars(req)
	varID := vars["id"]
	inventory := model.Inventory{Product: model.Product{Sku: varID}}

	// call service	
	res, err := h.workerService.GetInventory(ctx, &inventory)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusOK, res)
}

// About update inventory
func (h *HttpRouters) UpdateInventory(rw http.ResponseWriter, req *http.Request) error {
	ctx, cancel, span := h.withContext(req, "UpdateInventory")
	defer cancel()
	defer span.End()

	// decode payload	
	inventory := model.Inventory{}
	defer req.Body.Close()
	
	err := json.NewDecoder(req.Body).Decode(&inventory)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
	}

	// get put parameter		
	vars := mux.Vars(req)
	varSku := vars["id"]
	inventory.Product.Sku = varSku

	// call service	
	res, err := h.workerService.UpdateInventory(ctx, &inventory)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusOK, res)
}

// About get timeseries inventory data for a product
func (h *HttpRouters) ListInventory(rw http.ResponseWriter, req *http.Request) error {			
	ctx, cancel, span := h.withContext(req, "ListInventory")
	defer cancel()
	defer span.End()

	query := req.URL.Query()
	sku := query.Get("sku")
	if sku == "" {
		return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
	}

	// default window is 24, can be override by query parameter
	window := 10

	windowParam := query.Get("window")
	if windowParam != "" {
		parsedWindow, err := strconv.Atoi(windowParam)
		if err != nil || parsedWindow <= 0 {
			return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
		}
		window = parsedWindow
	}
	
	offset := 0
	offsetParam := query.Get("offset")
	if offsetParam != "" {
		parsedOffset, err := strconv.Atoi(offsetParam)
		if err != nil || parsedOffset < 0 {
			return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
		}
		offset = parsedOffset
	}

	inventory := model.Inventory{Product: model.Product{Sku: sku}}

	// call service	
	res, err := h.workerService.ListInventory(ctx, window, offset, &inventory)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusOK, res)
}