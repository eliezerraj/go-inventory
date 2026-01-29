package http

import (
	"time"
	"strconv"
	"net/http"
	"context"
	"strings"
	"encoding/json"	

	"github.com/rs/zerolog"
	"github.com/gorilla/mux"

	"github.com/go-inventory/shared/erro"
	"github.com/go-inventory/internal/domain/model"
	"github.com/go-inventory/internal/domain/service"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"

	go_core_midleware "github.com/eliezerraj/go-core/v2/middleware"
	go_core_otel_trace "github.com/eliezerraj/go-core/v2/otel/trace"
)

// Global middleware reference for error handling
var (
	_ go_core_midleware.MiddleWare
)

type HttpRouters struct {
	workerService 	*service.WorkerService
	appServer		*model.AppServer
	logger			*zerolog.Logger
	tracerProvider 	*go_core_otel_trace.TracerProvider
}

// Helper to extract context with timeout and setup span
func (h *HttpRouters) withContext(req *http.Request, spanName string) (context.Context, context.CancelFunc, trace.Span) {
	ctx, cancel := context.WithTimeout(req.Context(), 
		time.Duration(h.appServer.Server.CtxTimeout) * time.Second)
	
	h.logger.Info().
			Ctx(ctx).
			Str("func", spanName).Send()
	
	ctx, span := h.tracerProvider.SpanCtx(ctx, "adapter."+spanName, trace.SpanKindInternal)
	return ctx, cancel, span
}

// Helper to get trace ID from context using middleware function
func (h *HttpRouters) getTraceID(ctx context.Context) string {
	return go_core_midleware.GetRequestID(ctx)
}

// Above create routers
func NewHttpRouters(appServer *model.AppServer,
					workerService *service.WorkerService,
					appLogger *zerolog.Logger,
					tracerProvider *go_core_otel_trace.TracerProvider) HttpRouters {
	
	logger := appLogger.With().
				Str("package", "adapter.http").
				Logger()
			
	logger.Info().
			Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
		appServer: appServer,
		logger: &logger,
		tracerProvider: tracerProvider,
	}
}

// Helper to write JSON response
func (h *HttpRouters) writeJSON(w http.ResponseWriter, code int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	return json.NewEncoder(w).Encode(data)
}

// ErrorHandler creates an APIError with appropriate HTTP status based on error type
func (h *HttpRouters) ErrorHandler(traceID string, err error) *go_core_midleware.APIError {
	var httpStatusCode int = http.StatusInternalServerError

	if strings.Contains(err.Error(), "context deadline exceeded") {
		httpStatusCode = http.StatusGatewayTimeout
	}

	if strings.Contains(err.Error(), "check parameters") {
		httpStatusCode = http.StatusBadRequest
	}

	if strings.Contains(err.Error(), "not found") {
		httpStatusCode = http.StatusNotFound
	}

	if strings.Contains(err.Error(), "duplicate key") || 
	   strings.Contains(err.Error(), "unique constraint") {
		httpStatusCode = http.StatusBadRequest
	}

	return go_core_midleware.NewAPIError(err, traceID, httpStatusCode)
}

// About return a health, without log and trace to avoid flush then in K8 
func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About return a live, without log and trace to avoid flush then in K8 
func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About show all header received
func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	h.logger.Info().
			Str("func","Header").Send()
	
	json.NewEncoder(rw).Encode(req.Header)
}

// About show all context values
func (h *HttpRouters) Context(rw http.ResponseWriter, req *http.Request) {
	h.logger.Info().
			Str("func","Context").Send()
	
	json.NewEncoder(rw).Encode(req.Context())
}

// About info
func (h *HttpRouters) Info(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	_, cancel, span := h.withContext(req, "Info")
	defer cancel()
	defer span.End()

	json.NewEncoder(rw).Encode(h.appServer)
}

// About add product
func (h *HttpRouters) AddProduct(rw http.ResponseWriter, req *http.Request) error {
	ctx, cancel, span := h.withContext(req, "AddProduct")
	defer cancel()
	defer span.End()
	
	// decode payload		
	product := model.Product{}
	defer req.Body.Close()
	
	err := json.NewDecoder(req.Body).Decode(&product)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())		
		return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
	}

	// call service
	res, err := h.workerService.AddProduct(ctx, &product)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusOK, res)
}

// About get product
func (h *HttpRouters) GetProduct(rw http.ResponseWriter, req *http.Request) error {
	ctx, cancel, span := h.withContext(req, "GetProduct")
	defer cancel()
	defer span.End()

    // Debug: Check if request ID is in context
    //requestID := go_core_midleware.GetRequestID(ctx)
    //if requestID == "" {
    //    h.logger.Warn().Msg("Request ID is empty in context")
    //    // Call debug helper to investigate
    //    go_core_midleware.DebugContextValues(ctx, h.logger)
    //}

	  // For demonstration; remove in production
	// decode payload				
	vars := mux.Vars(req)
	varID := vars["id"]
	product := model.Product{Sku: varID}
	
	// call service	
	res, err := h.workerService.GetProduct(ctx, &product)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusOK, res)
}

// About get product by ID
func (h *HttpRouters) GetProductId(rw http.ResponseWriter, req *http.Request) error {
	ctx, cancel, span := h.withContext(req, "GetProductId")
	defer cancel()
	defer span.End()

	// decode payload				
	vars := mux.Vars(req)
	varID := vars["id"]
	
	varIDint, err := strconv.Atoi(varID)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		return h.ErrorHandler(h.getTraceID(ctx), erro.ErrBadRequest)
	}

	product := model.Product{ID: varIDint}
	
	// call service	
	res, err := h.workerService.GetProductId(ctx, &product)
	if err != nil {
		return h.ErrorHandler(h.getTraceID(ctx), err)
	}
	
	return h.writeJSON(rw, http.StatusOK, res)
}

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
