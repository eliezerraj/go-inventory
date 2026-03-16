package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/go-inventory/internal/domain/model"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// Helper function to scan inventory from rows iterator
func (w *WorkerRepository) scanInventoryFromRows(rows pgx.Rows) (*model.Inventory, error) {
	inventory := model.Inventory{}
	product := model.Product{}

	err := rows.Scan(
					&product.Sku,
					&product.ID,
					&inventory.CreatedAt,
					&inventory.Available,
					&inventory.Sold,
					&inventory.Pending,
					&inventory.Incoming,
				)
	if err != nil {
		return nil, fmt.Errorf("FAILED to scan inventory from rows: %w", err)
	}
	// Set product
	inventory.Product = product
	
	return &inventory, nil
}

// AddInventoryTimeSeries inserts a new inventory record into the inventory_time_series table and returns the inserted inventory with its ID.
func (w* WorkerRepository) AddInventoryTimeSeries(ctx context.Context, 
												tx pgx.Tx, 
												inventory *model.Inventory) (*model.Inventory, error){
	w.logger.Info().
			Ctx(ctx).
			Str("func","AddInventoryTimeSeries").Send()

	// trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.AddInventoryTimeSeries", trace.SpanKindInternal)
	defer span.End()

	//Prepare
	var id int

	// Query Execute
	query := `INSERT INTO inventory_time_series ( 	snapshot_date,
													fk_product_id,
													available,
													pending,
													sold,
													incoming,
													created_at) 
				VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	row := tx.QueryRow(	ctx, 
						query,
						inventory.CreatedAt,
						inventory.Product.ID,
						inventory.Available, 
						inventory.Pending,
						inventory.Sold,
						inventory.Incoming,
						inventory.CreatedAt)
						
	if err := row.Scan(&id); err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, fmt.Errorf("FAILED to insert inventory_time_series: %w", err)
	}

	// Set PK
	inventory.ID = id
	
	return inventory , nil
}

// About get a product
func (w *WorkerRepository) GetInventoryTimeSeries(	ctx context.Context, 
													windowsize int,
									  				inventory *model.Inventory)  (*[]model.Inventory, error){
	w.logger.Info().
			Ctx(ctx).
			Str("func","GetInventoryTimeSeries").Send()

	// Trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.GetInventoryTimeSeries", trace.SpanKindInternal)
	defer span.End()

	// db connection
	conn, err := w.DatabasePG.Acquire(ctx)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		w.logger.Error().
			  	 Ctx(ctx).
				 Err(err).Send()
		return nil, fmt.Errorf("FAILED to acquire connection: %w", err)
	}
	defer w.DatabasePG.Release(conn)

	// Query and Execute
	query := `select 	pr.sku,
						its.fk_product_id, 
						its.snapshot_date, 
						its.available,
						its.sold,
						its.pending,
						its.incoming
				from inventory_time_series its,
						product pr
				where pr.sku = $1
				and its.fk_product_id = pr.id
				order by its.fk_product_id desc, its.snapshot_date desc
				limit $2`

	rows, err := conn.Query(ctx, 
							query, 
							inventory.Product.Sku, 
							windowsize,)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, fmt.Errorf("FAILED to query inventory_time_series: %w", err)
	}
	defer rows.Close()

	list_inventory := []model.Inventory{}
	for rows.Next() {
		res_inventory, err := w.scanInventoryFromRows(rows)
		if err != nil {
			span.RecordError(err) 
        	span.SetStatus(codes.Error, err.Error())
			w.logger.Error().
					Ctx(ctx).
					Err(err).Send()
			return nil, fmt.Errorf("FAILED to scanInventoryFromRows inventory_time_series: %w", err)
		}
		list_inventory = append(list_inventory, *res_inventory)
	}

	return &list_inventory, nil
}
