package database

import (
	"context"
	"fmt"
	"database/sql"

	"github.com/jackc/pgx/v5"
	
	"github.com/go-inventory/shared/erro"
	"github.com/go-inventory/internal/domain/model"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

//--------------------------------------
// About create a Inventory
func (w* WorkerRepository) AddInventory(ctx context.Context, 
										tx pgx.Tx, 
										inventory *model.Inventory) (*model.Inventory, error){
	w.logger.Info().
			Ctx(ctx).
			Str("func","AddInventory").Send()

	// trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.AddInventory", trace.SpanKindInternal)
	defer span.End()

	//Prepare
	var id int

	// Query Execute
	query := `INSERT INTO inventory ( 	fk_product_id,
										available,
										pending,
										reserved,
										sold,
										created_at) 
				VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	row := tx.QueryRow(	ctx, 
						query,
						inventory.Product.ID,
						inventory.Available, 
						inventory.Pending,
						inventory.Reserved,
						inventory.Sold,
						inventory.CreatedAt)
						
	if err := row.Scan(&id); err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, fmt.Errorf("FAILED to insert inventory: %w", err)
	}

	// Set PK
	inventory.ID = id
	
	return inventory , nil
}

// About get a Inventory
func (w *WorkerRepository) GetInventory(ctx context.Context, 
										inventory *model.Inventory) (*model.Inventory, error){
	w.logger.Info().
			Ctx(ctx).
			Str("func","GetInventory").Send()
			
	// Trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.GetInventory", trace.SpanKindInternal)
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

	// Prepare
	res_product := model.Product{}
	res_inventory := model.Inventory{}
	var nullProductUpdatedAt sql.NullTime
	var nullInventoryUpdatedAt sql.NullTime

	// Query and Execute
	query := `SELECT p.id, 
					 p.sku, 
					 p.type,
					 p.name,
					 p.status,
					 p.lead_time,
					 p.created_at, 
					 p.updated_at,
					 i.id,
					 i.available,
					 i.pending,
					 i.reserved,
					 i.sold,
					 i.created_at,
					 i.updated_at
				FROM product as p,
					 inventory as i
				WHERE sku =$1
				and p.id = i.fk_product_id`

	rows, err := conn.Query(ctx, 
							query, 
							inventory.Product.Sku)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())		
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, fmt.Errorf("FAILED to query inventory: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&res_product.ID, 
						&res_product.Sku, 
						&res_product.Type,
						&res_product.Name,
						&res_product.Status,
						&res_product.LeadTime,
						&res_product.CreatedAt,
						&nullProductUpdatedAt,
						&res_inventory.ID, 
						&res_inventory.Available, 
						&res_inventory.Pending,
						&res_inventory.Reserved, 
						&res_inventory.Sold,
						&res_inventory.CreatedAt,
						&nullInventoryUpdatedAt,
					)
		if err != nil {
			span.RecordError(err) 
        	span.SetStatus(codes.Error, err.Error())
			w.logger.Error().
					Ctx(ctx).
					Err(err).Send()
			return nil, fmt.Errorf("FAILED to scan inventory row: %w", err)
		}

		res_product.UpdatedAt = w.pointerTime(nullProductUpdatedAt)
		res_inventory.UpdatedAt = w.pointerTime(nullInventoryUpdatedAt)
		res_inventory.Product = res_product
		return &res_inventory, nil
	}

	w.logger.Warn().
			Ctx(ctx).
			Err(erro.ErrNotFound).Send()

	return nil, erro.ErrNotFound
}

// About update a Inventory
func (w* WorkerRepository) UpdateInventory(ctx context.Context, 
											tx pgx.Tx, 
											inventory *model.Inventory) (int64, error){

	w.logger.Info().
			Ctx(ctx).
			Str("func","UpdateInventory").Send()

	// trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.UpdateInventory", trace.SpanKindInternal)
	defer span.End()

	// Query Execute
	query := `UPDATE inventory
				SET available = available + $3,
					reserved = reserved + $4,
					pending = pending + $6,
					sold = sold + $5,
					updated_at = $2
				WHERE id = (SELECT id 
							FROM inventory
							WHERE fk_product_id = $1
							ORDER BY id
							FOR UPDATE SKIP LOCKED 
							LIMIT 1)`

	row, err := tx.Exec(ctx, 
						query,	
						inventory.Product.ID,
						inventory.UpdatedAt,		
						inventory.Available,
						inventory.Reserved,
						inventory.Sold,
						inventory.Pending,
					)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return 0, fmt.Errorf("FAILED to update inventory: %w", err)
	}

	return row.RowsAffected(), nil
}

// About get a product
func (w *WorkerRepository) ListInventory(ctx context.Context, 
										 limit int,
										 offset int,
									  	 inventory *model.Inventory)  (*[]model.Inventory, error){
	w.logger.Info().
			Ctx(ctx).
			Str("func","ListInventory").Send()

	// Trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.ListInventory", trace.SpanKindInternal)
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
	query := `SELECT p.id, 
					 p.sku, 
					 p.type,
					 p.name,
					 p.status,
					 p.lead_time,
					 p.created_at, 
					 p.updated_at,
					 i.id,
					 i.available,
					 i.pending,
					 i.reserved,
					 i.sold,
					 i.created_at,
					 i.updated_at
				FROM product as p,
					 inventory as i
				WHERE p.id = i.fk_product_id
				and p.sku like '%' || $1 || '%'
				order by p.sku asc
				limit $2 offset $3;`

	rows, err := conn.Query(ctx, 
							query, 
							inventory.Product.Sku, 
							limit,
							offset)
	if err != nil {
		span.RecordError(err) 
        span.SetStatus(codes.Error, err.Error())
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, fmt.Errorf("FAILED to query inventory: %w", err)
	}
	defer rows.Close()

	list_inventory := []model.Inventory{}

	var nullProductUpdatedAt sql.NullTime
	var nullInventoryUpdatedAt sql.NullTime

	for rows.Next() {
		res_product := model.Product{}
		res_inventory := model.Inventory{}

		err := rows.Scan(&res_product.ID, 
						&res_product.Sku, 
						&res_product.Type,
						&res_product.Name,
						&res_product.Status,
						&res_product.LeadTime,
						&res_product.CreatedAt,
						&nullProductUpdatedAt,
						&res_inventory.ID, 
						&res_inventory.Available, 
						&res_inventory.Pending,
						&res_inventory.Reserved, 
						&res_inventory.Sold,
						&res_inventory.CreatedAt,
						&nullInventoryUpdatedAt,
					)
		if err != nil {
			span.RecordError(err) 
        	span.SetStatus(codes.Error, err.Error())
			w.logger.Error().
					Ctx(ctx).
					Err(err).Send()
			return nil, fmt.Errorf("FAILED to scan inventory row: %w", err)
		}

		res_product.UpdatedAt = w.pointerTime(nullProductUpdatedAt)
		res_inventory.UpdatedAt = w.pointerTime(nullInventoryUpdatedAt)
		res_inventory.Product = res_product

		list_inventory = append(list_inventory, res_inventory)
	}
	
	if len(list_inventory) > 0 {
		return &list_inventory, nil
	}

	return nil, erro.ErrNotFound
}