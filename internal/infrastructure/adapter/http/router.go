package http

import (
	"fmt"
	"time"
	"reflect"
	"net/http"
	"context"
	"strings"
	"encoding/json"	

	"github.com/rs/zerolog"
	"github.com/gorilla/mux"

	"github.com/go-inventory/shared/erro"
	"github.com/go-inventory/internal/domain/model"
	"github.com/go-inventory/internal/domain/service"

	go_core_json "github.com/eliezerraj/go-core/coreJson"
	go_core_otel_trace "github.com/eliezerraj/go-core/otel/trace"
)

var (
	coreJson 		go_core_json.CoreJson
	coreApiError 	go_core_json.APIError
	tracerProvider go_core_otel_trace.TracerProvider
)

type HttpRouters struct {
	workerService 	*service.WorkerService
	appServer		*model.AppServer
	logger			*zerolog.Logger
}

// Type for async result
type result struct {
		data interface{}
		err  error
}

// Above create routers
func NewHttpRouters(appServer *model.AppServer,
					workerService *service.WorkerService,
					appLogger *zerolog.Logger) HttpRouters {
	logger := appLogger.With().
						Str("package", "adapter.http").
						Logger()

	logger.Info().
			Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
		appServer: appServer,
		logger: &logger,
	}
}

// About handle error
func (h *HttpRouters) ErrorHandler(trace_id string, err error) *go_core_json.APIError {

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

	if strings.Contains(err.Error(), "INSERT has more target") {
    	httpStatusCode = http.StatusInternalServerError
	}

	coreApiError = coreApiError.NewAPIError(err, trace_id, httpStatusCode)

	return &coreApiError
}

// About return a health
func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	h.logger.Info().
			Str("func","Health").Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About return a live
func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	h.logger.Info().
			Str("func","Live").Send()

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
	
	contextValues := reflect.ValueOf(req.Context()).Elem()

	json.NewEncoder(rw).Encode(fmt.Sprintf("%v",contextValues))
}

// About info
func (h *HttpRouters) Info(rw http.ResponseWriter, req *http.Request) {
	// extract context		
	ctx, cancel := context.WithTimeout(req.Context(), 
										time.Duration(h.appServer.Server.CtxTimeout) * time.Second)
    defer cancel()

	// trace	
	ctx, span := tracerProvider.SpanCtx(ctx, "adapter.http.Info")
	defer span.End()

	// log with context
	h.logger.Info().
			Ctx(ctx).
			Str("func","Info").Send()

	json.NewEncoder(rw).Encode(h.appServer)
}

// About add product
func (h *HttpRouters) AddProduct(rw http.ResponseWriter, req *http.Request) error {
	// extract context	
	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(h.appServer.Server.CtxTimeout) * time.Second)
    defer cancel()

	// trace	
	ctx, span := tracerProvider.SpanCtx(ctx, "adapter.http.AddProduct")
	defer span.End()
	
	// log with context
	h.logger.Info().
			Ctx(ctx).
			Str("func","AddProduct").Send()

	product := model.Product{}
	
	err := json.NewDecoder(req.Body).Decode(&product)
    if err != nil {
		trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
		return h.ErrorHandler(trace_id, erro.ErrBadRequest)
    }
	defer req.Body.Close()

	res, err := h.workerService.AddProduct(ctx, &product)
	if err != nil {
		trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
		return h.ErrorHandler(trace_id, err)
	}
	
	return coreJson.WriteJSON(rw, http.StatusOK, res)
}

// About get product
func (h *HttpRouters) GetProduct(rw http.ResponseWriter, req *http.Request) error {
	// extract context		
	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(h.appServer.Server.CtxTimeout) * time.Second)
    defer cancel()

	// trace	
	ctx, span := tracerProvider.SpanCtx(ctx, "adapter.http.GetProduct")
	defer span.End()

	// log with context
	h.logger.Info().
			Ctx(ctx).
			Str("func","GetProduct").Send()

	vars := mux.Vars(req)
	varID := vars["id"]

	product := model.Product{Sku: varID}
	
	res, err := h.workerService.GetProduct(ctx, &product)
	if err != nil {
		trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
		return h.ErrorHandler(trace_id, err)
	}
	
	return coreJson.WriteJSON(rw, http.StatusOK, res)
}

// About get inventory
func (h *HttpRouters) GetInventory(rw http.ResponseWriter, req *http.Request) error {
	// extract context		
	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(h.appServer.Server.CtxTimeout) * time.Second)
    defer cancel()

	// trace	
	ctx, span := tracerProvider.SpanCtx(ctx, "adapter.http.GetInventory")
	defer span.End()

	// log with context
	h.logger.Info().
			Ctx(ctx).
			Str("func","GetInventory").Send()

	vars := mux.Vars(req)
	varID := vars["id"]

	inventory := model.Inventory{ 
									Product: model.Product{Sku: varID} ,
								}

	res, err := h.workerService.GetInventory(ctx, &inventory)
	if err != nil {
		trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
		return h.ErrorHandler(trace_id, err)
	}
	
	return coreJson.WriteJSON(rw, http.StatusOK, res)
}
