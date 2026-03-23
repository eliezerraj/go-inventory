package service

import (
	"time"
	"context"

	"github.com/go-inventory/internal/domain/model"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

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
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
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
	resInventory.Pending = inventory.Pending + resInventory.Pending
	resInventory.Sold = inventory.Sold + resInventory.Sold
	resInventory.UpdatedAt = inventory.UpdatedAt	

	// create a time series for inventory only for sold products (order checkout)
	inventory.Available = resInventory.Available
	inventory.CreatedAt = time.Now()
	// pending can be negative when an order is completed, but in the time series we want to keep it as zero to avoid confusion in the reports
	if inventory.Pending < 0 {
		inventory.Pending = 0
	}
	_, err = s.workerRepository.AddInventoryTimeSeries(ctx, tx, inventory)
	if err != nil {
		return nil, err
	}

	return resInventory, nil
}

// About get inventory
func (s * WorkerService) ListInventory(ctx context.Context, limit int, offset int, inventory *model.Inventory) (*[]model.Inventory, error){
	result, err := s.callRepositoryRead(ctx, "ListInventory", func(ctx context.Context) (interface{}, error) {
		return s.workerRepository.ListInventory(ctx, limit, offset, inventory)
	})
	
	if err != nil {
		return nil, err
	}
	return result.(*[]model.Inventory), nil
}
