package service

import (
	"time"
	"context"

	"github.com/rs/zerolog"

	"github.com/go-inventory/internal/domain/model"
	"github.com/go-inventory/shared/erro"

	database "github.com/go-inventory/internal/infrastructure/repo/database"

	go_core_db_pg "github.com/eliezerraj/go-core/v2/database/postgre"
	go_core_otel_trace "github.com/eliezerraj/go-core/v2/otel/trace"
)

var tracerProvider go_core_otel_trace.TracerProvider

type WorkerService struct {
	workerRepository *database.WorkerRepository
	logger 			*zerolog.Logger
}

// About new worker service
func NewWorkerService(	workerRepository *database.WorkerRepository, 
						appLogger *zerolog.Logger) *WorkerService{
	logger := appLogger.With().
						Str("package", "domain.service").
						Logger()
	logger.Info().
			Str("func","NewWorkerService").Send()

	return &WorkerService{
		workerRepository: workerRepository,
		logger: &logger,
	}
}

// About check health service
func (s * WorkerService) HealthCheck(ctx context.Context) error{
	s.logger.Info().
			Ctx(ctx).
			Str("func","HealthCheck").Send()

	ctx, span := tracerProvider.SpanCtx(ctx, "service.HealthCheck")
	defer span.End()

	// Check database health
	_, spanDB := tracerProvider.SpanCtx(ctx, "DatabasePG.Ping")
	err := s.workerRepository.DatabasePG.Ping()
	if err != nil {
		s.logger.Error().
				Ctx(ctx).
				Err(err).Msg("*** Database HEALTH CHECK FAILED ***")
		return erro.ErrHealthCheck
	}
	spanDB.End()

	s.logger.Info().
			Ctx(ctx).
			Str("func","HealthCheck").
			Msg("*** Database HEALTH CHECK SUCCESSFULL ***")

	return nil
}

// About database stats
func (s *WorkerService) Stat(ctx context.Context) (go_core_db_pg.PoolStats){
	s.logger.Info().
			Ctx(ctx).
			Str("func","Stat").Send()

	return s.workerRepository.Stat(ctx)
}

// About create a product
func (s *WorkerService) AddProduct(ctx context.Context, 
									product *model.Product) (*model.Inventory, error){
	s.logger.Info().
			Ctx(ctx).
			Str("func","AddProduct").Send()
	// trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.AddProduct")

	// prepare database
	tx, conn, err := s.workerRepository.DatabasePG.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	// handle connection
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePG.ReleaseTx(conn)
		span.End()
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
	s.logger.Info().
			Ctx(ctx).
			Str("func","GetProduct").Send()

	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.GetProduct")
	defer span.End()

	// Call a service
	res, err := s.workerRepository.GetProduct(ctx, product)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// About get a product
func (s * WorkerService) GetProductId(ctx context.Context, product *model.Product) (*model.Product, error){
	s.logger.Info().
			Ctx(ctx).
			Str("func","GetProductId").Send()

	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.GetProductId")
	defer span.End()

	// Call a service
	res, err := s.workerRepository.GetProductId(ctx, product)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// About get inventory
func (s * WorkerService) GetInventory(ctx context.Context, inventory *model.Inventory) (*model.Inventory, error){
	s.logger.Info().
			Ctx(ctx).
			Str("func","GetInventory").Send()

	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.GetInventory")
	defer span.End()
	
	// Call a service
	res, err := s.workerRepository.GetInventory(ctx, inventory)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// About get inventory
func (s * WorkerService) UpdateInventory(ctx context.Context, inventory *model.Inventory) (*model.Inventory, error){
	s.logger.Info().
			Ctx(ctx).
			Str("func","UpdateInventory").Send()
	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.UpdateInventory")
	
	// prepare database
	tx, conn, err := s.workerRepository.DatabasePG.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	// handle connection
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePG.ReleaseTx(conn)
		span.End()
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
	
	// whenever zero rows was update, for the skip lock clausule, a new rows must be inserted
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
