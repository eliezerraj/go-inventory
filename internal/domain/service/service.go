package service

import (
	"time"
	"context"

	"github.com/rs/zerolog"

	"github.com/go-inventory/internal/domain/model"
	"github.com/go-inventory/shared/erro"

	database "github.com/go-inventory/internal/infrastructure/repo/database"
	go_core_db_pg "github.com/eliezerraj/go-core/database/postgre"
	go_core_otel_trace "github.com/eliezerraj/go-core/otel/trace"
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
						Str("component", "domain.service").
						Logger()
	logger.Debug().
			Str("func","NewDatabasePGServer").Send()

	return &WorkerService{
		workerRepository: workerRepository,
		logger: &logger,
	}
}

func (s *WorkerService) Stat(ctx context.Context) (go_core_db_pg.PoolStats){
	s.logger.Info().
			Str("func","Stat").Send()

	return s.workerRepository.Stat(ctx)
}

// About create a product
func (s *WorkerService) AddProduct(ctx context.Context, 
									product *model.Product) (*model.Inventory, error){
	// trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.AddProduct")
	defer span.End()

	s.logger.Info().
			Ctx(ctx).
			Str("func","AddProduct").Send()

	// prepare database
	tx, conn, err := s.workerRepository.DatabasePG.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.workerRepository.DatabasePG.ReleaseTx(conn)

	// handle connection
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		span.End()
	}()

	// prepare data
	product.CreatedAt = time.Now()

	// Create product
	res_product, err := s.workerRepository.AddProduct(ctx, tx, product)
	if err != nil {
		return nil, err
	}

	// Setting PK
	product.ID = res_product.ID

	// Prepate inventory
	inventory := model.Inventory{
		Product: 		*res_product,
		QtdAvailable:	1000,
		QtdReserved:	0,
		QtdTotal:		1000, 	
		CreatedAt:		time.Now(),
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
	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.GetProduct")
	defer span.End()

	// log with context
	s.logger.Info().
			Ctx(ctx).
			Str("func","GetProduct").Send()

	// Call a service
	res, err := s.workerRepository.GetProduct(ctx, product)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// About check health service
func (s * WorkerService) HealthCheck(ctx context.Context) error{
	s.logger.Info().
			Str("func","HealthCheck").Send()

	// Check database health
	err := s.workerRepository.DatabasePG.Ping()
	if err != nil {
		s.logger.Error().
				Err(err).Msg("*** Database HEALTH FAILED ***")
		return erro.ErrHealthCheck
	}

	s.logger.Info().
			Str("func","HealthCheck").
			Msg("*** Database HEALTH SUCCESSFULL ***")

	return nil
}

// About get a product
func (s * WorkerService) GetInventory(ctx context.Context, inventory *model.Inventory) (*model.Inventory, error){
	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.GetInventory")
	defer span.End()
	
	s.logger.Info().
			Ctx(ctx).
			Str("func","GetInventory").Send()

	// Call a service
	res, err := s.workerRepository.GetInventory(ctx, inventory)
	if err != nil {
		return nil, err
	}

	return res, nil
}
