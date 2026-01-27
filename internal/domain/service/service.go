package service

import (
	"time"
	"context"

	"github.com/rs/zerolog"

	"github.com/go-inventory/internal/domain/model"
	"github.com/go-inventory/shared/erro"
	"go.opentelemetry.io/otel/trace"

	database "github.com/go-inventory/internal/infrastructure/repo/database"

	go_core_db_pg "github.com/eliezerraj/go-core/v2/database/postgre"
	go_core_otel_trace "github.com/eliezerraj/go-core/v2/otel/trace"
)

type WorkerService struct {
	workerRepository *database.WorkerRepository
	logger 			*zerolog.Logger
	tracerProvider 	*go_core_otel_trace.TracerProvider
}

// About new worker service
func NewWorkerService(	workerRepository *database.WorkerRepository, 
						appLogger 		*zerolog.Logger,
						tracerProvider 	*go_core_otel_trace.TracerProvider) *WorkerService{
							
	logger := appLogger.With().
						Str("package", "domain.service").
						Logger()
	logger.Info().
			Str("func","NewWorkerService").Send()

	return &WorkerService{
		workerRepository: workerRepository,
		logger: &logger,
		tracerProvider: tracerProvider,
	}
}

// Helper function for common repository read operations
func (s *WorkerService) callRepositoryRead(ctx context.Context, spanName string, 
	fn func(context.Context) (interface{}, error)) (interface{}, error) {
	s.logger.Info().
			Ctx(ctx).
			Str("func", spanName).Send()
	
	ctx, span := s.tracerProvider.SpanCtx(ctx, "service."+spanName, trace.SpanKindServer)
	defer span.End()
	
	return fn(ctx)
}


// About database stats
func (s *WorkerService) Stat(ctx context.Context) (go_core_db_pg.PoolStats){
	s.logger.Info().
			Ctx(ctx).
			Str("func","Stat").Send()

	return s.workerRepository.Stat(ctx)
}

// About check health service
func (s * WorkerService) HealthCheck(ctx context.Context) error{
	s.logger.Info().
			Ctx(ctx).
			Str("func","HealthCheck").Send()

	ctx, span := s.tracerProvider.SpanCtx(ctx, "service.HealthCheck", trace.SpanKindServer)
	defer span.End()

	// Check database health
	ctx, spanDB := s.tracerProvider.SpanCtx(ctx, "DatabasePG.Ping", trace.SpanKindClient)
	err := s.workerRepository.DatabasePG.Ping()
	spanDB.End()
	
	if err != nil {
		s.logger.Error().
				Ctx(ctx).
				Err(err).Msg("*** Database HEALTH CHECK FAILED ***")
		return erro.ErrHealthCheck
	}

	s.logger.Info().
			Ctx(ctx).
			Str("func","HealthCheck").
			Msg("*** Database HEALTH CHECK SUCCESSFULL ***")

	return nil
}

// About create a product
func (s *WorkerService) AddProduct(ctx context.Context, 
									product *model.Product) (*model.Inventory, error){
	s.logger.Info().
			Ctx(ctx).
			Str("func","AddProduct").Send()
	// trace
	ctx, span := s.tracerProvider.SpanCtx(ctx, "service.AddProduct", trace.SpanKindServer)
	defer span.End()

	// prepare database
	tx, conn, err := s.workerRepository.DatabasePG.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	// handle connection and transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePG.ReleaseTx(conn)
	}()

	// prepare data
	now := time.Now()
	product.CreatedAt = now

	// Create product
	res_product, err := s.workerRepository.AddProduct(ctx, tx, product)
	if err != nil {
		return nil, err
	}

	// Setting PK
	product.ID = res_product.ID

	// Create a default inventory
	inventory := model.Inventory{
		Product: 	*res_product,
		Available:	1000,
		Reserved:	0,
		Sold:		0, 	
		CreatedAt:	now,
	}
	//Create inventory
	res_inventory, err := s.workerRepository.AddInventory(ctx, tx, &inventory)
	if err != nil {
		return nil, err
	}

	return res_inventory, nil
}

// About get a product
func (s * WorkerService) GetProduct(ctx context.Context, product *model.Product) (*model.Product, error){
	result, err := s.callRepositoryRead(ctx, "GetProduct", func(ctx context.Context) (interface{}, error) {
		return s.workerRepository.GetProduct(ctx, product)
	})
	if err != nil {
		return nil, err
	}
	return result.(*model.Product), nil
}

// About get a product by ID
func (s * WorkerService) GetProductId(ctx context.Context, product *model.Product) (*model.Product, error){
	result, err := s.callRepositoryRead(ctx, "GetProductId", func(ctx context.Context) (interface{}, error) {
		return s.workerRepository.GetProductId(ctx, product)
	})
	if err != nil {
		return nil, err
	}
	return result.(*model.Product), nil
}

// About get inventory
func (s * WorkerService) GetInventory(ctx context.Context, inventory *model.Inventory) (*model.Inventory, error){
	result, err := s.callRepositoryRead(ctx, "GetInventory", func(ctx context.Context) (interface{}, error) {
		return s.workerRepository.GetInventory(ctx, inventory)
	})
	if err != nil {
		return nil, err
	}
	return result.(*model.Inventory), nil
}

// About update inventory
func (s * WorkerService) UpdateInventory(ctx context.Context, inventory *model.Inventory) (*model.Inventory, error){
	s.logger.Info().
			Ctx(ctx).
			Str("func","UpdateInventory").Send()
	
	// Trace
	ctx, span := s.tracerProvider.SpanCtx(ctx, "service.UpdateInventory", trace.SpanKindServer)
	defer span.End()
	
	// prepare database
	tx, conn, err := s.workerRepository.DatabasePG.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	// handle connection and transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePG.ReleaseTx(conn)
	}()

	// Get product info
	resInventory, err := s.workerRepository.GetInventory(ctx, inventory)
	if err != nil {
		return nil, err
	}

	// set data for update
	now := time.Now()
	inventory.UpdatedAt = &now
	inventory.ID = resInventory.ID
	inventory.Product = resInventory.Product

	// Call a service
	row, err := s.workerRepository.UpdateInventory(ctx, tx, inventory)
	if err != nil {
		return nil, err
	}
	
	// whenever zero rows was updated, due to the skip lock clause, a new row must be inserted
	if row == 0 {
		_, err := s.workerRepository.AddInventory(ctx, tx, resInventory)
		if err != nil {
			return nil, err
		}
	}

	resInventory.Available = inventory.Available + resInventory.Available
	resInventory.Reserved = inventory.Reserved + resInventory.Reserved
	resInventory.Sold = inventory.Sold + resInventory.Sold
	resInventory.UpdatedAt = inventory.UpdatedAt	
	
	return resInventory, nil
}
