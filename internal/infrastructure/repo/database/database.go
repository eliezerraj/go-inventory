package database

import (
	"context"
	"fmt"
	"strings"
	"database/sql"
	"time"

	"github.com/rs/zerolog"
	"github.com/jackc/pgx/v5"

	"github.com/go-inventory/shared/erro"
	"github.com/go-inventory/internal/domain/model"
	"go.opentelemetry.io/otel/trace"

	go_core_otel_trace "github.com/eliezerraj/go-core/v2/otel/trace"
	go_core_db_pg "github.com/eliezerraj/go-core/v2/database/postgre"
)

type WorkerRepository struct {
	DatabasePG 		*go_core_db_pg.DatabasePGServer
	logger			*zerolog.Logger
	tracerProvider 	*go_core_otel_trace.TracerProvider
}

// Above new worker
func NewWorkerRepository(databasePG *go_core_db_pg.DatabasePGServer,
						appLogger *zerolog.Logger,
						tracerProvider *go_core_otel_trace.TracerProvider) *WorkerRepository{
	logger := appLogger.With().
						Str("package", "repo.database").
						Logger()
	logger.Info().
			Str("func","NewWorkerRepository").Send()

	return &WorkerRepository{
		DatabasePG: databasePG,
		logger: &logger,
		tracerProvider: tracerProvider,
	}
}

// Helper function to convert nullable time to pointer
func (w *WorkerRepository) pointerTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// Helper function to scan product from rows iterator
func (w *WorkerRepository) scanProductFromRows(rows pgx.Rows) (*model.Product, error) {
	product := model.Product{}
	var nullUpdatedAt sql.NullTime
	
	err := rows.Scan(&product.ID, 
					&product.Sku, 
					&product.Type,
					&product.Name,
					&product.Status,
					&product.CreatedAt,
					&nullUpdatedAt,
				)
	if err != nil {
		return nil, fmt.Errorf("FAILED to scan product from rows: %w", err)
	}
	
	product.UpdatedAt = w.pointerTime(nullUpdatedAt)
	return &product, nil
}

// Above get stats from database
func (w *WorkerRepository) Stat(ctx context.Context) (go_core_db_pg.PoolStats){
	w.logger.Info().
			Ctx(ctx).
			Str("func","Stat").Send()
	
	stats := w.DatabasePG.Stat()

	resPoolStats := go_core_db_pg.PoolStats{
		AcquireCount:         stats.AcquireCount(),
		AcquiredConns:        stats.AcquiredConns(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		ConstructingConns:    stats.ConstructingConns(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
		IdleConns:            stats.IdleConns(),
		MaxConns:             stats.MaxConns(),
		TotalConns:           stats.TotalConns(),
	}

	return resPoolStats
}

// About create a product
func (w* WorkerRepository) AddProduct(ctx context.Context, 
									tx pgx.Tx, 
									product *model.Product) (*model.Product, error){
	w.logger.Info().
			Ctx(ctx).
			Str("func","AddProduct").Send()
			
	// trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.AddProduct", trace.SpanKindInternal)
	defer span.End()

	//Prepare
	var id int

	// Query Execute
	query := `INSERT INTO product ( sku, 
									type,
									name,
									status,
									created_at) 
				VALUES($1, $2, $3, $4, $5) RETURNING id`

	row := tx.QueryRow(	ctx, 
						query,
						product.Sku,
						product.Type, 
						product.Name,
						product.Status,
						product.CreatedAt)
						
	if err := row.Scan(&id); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates") {
    		w.logger.Warn().
					 Ctx(ctx).
					 Err(err).Send()
		} else {
			w.logger.Error().
					 Ctx(ctx).
				     Err(err).Send()
		}
		return nil, fmt.Errorf("FAILED to insert product: %w", err)
	}

	// Set PK
	product.ID = id
	
	return product , nil
}

// About get a product
func (w *WorkerRepository) GetProduct(ctx context.Context, 
									  product *model.Product) (*model.Product, error){
	w.logger.Info().
			Ctx(ctx).
			Str("func","GetProduct").Send()

	// Trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.GetProduct", trace.SpanKindInternal)
	defer span.End()

	// db connection
	conn, err := w.DatabasePG.Acquire(ctx)
	if err != nil {
		w.logger.Error().
			  	 Ctx(ctx).
				 Err(err).Send()
		return nil, fmt.Errorf("FAILED to acquire connection: %w", err)
	}
	defer w.DatabasePG.Release(conn)

	// Query and Execute
	query := `SELECT id, 
					sku, 
					type,
					name,
					status,
					created_at, 
					updated_at
				FROM product 
				WHERE sku =$1`

	rows, err := conn.Query(ctx, 
							query, 
							product.Sku)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, fmt.Errorf("FAILED to query product: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		res_product, err := w.scanProductFromRows(rows)
		if err != nil {
			w.logger.Error().
					Ctx(ctx).
					Err(err).Send()
			return nil, err
		}
		return res_product, nil
	}

	w.logger.Warn().
			Ctx(ctx).
			Err(erro.ErrNotFound).Send()

	return nil, erro.ErrNotFound
}

// About get a product by ID
func (w *WorkerRepository) GetProductId(ctx context.Context, 
										product *model.Product) (*model.Product, error){
	w.logger.Info().
			Ctx(ctx).
			Str("func","GetProductId").Send()

	// Trace
	ctx, span := w.tracerProvider.SpanCtx(ctx, "database.GetProductId", trace.SpanKindInternal)
	defer span.End()

	// db connection
	conn, err := w.DatabasePG.Acquire(ctx)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, fmt.Errorf("FAILED to acquire connection: %w", err)
	}
	defer w.DatabasePG.Release(conn)

	// Query and Execute
	query := `SELECT id, 
					sku, 
					type,
					name,
					status,
					created_at, 
					updated_at
				FROM product 
				WHERE id =$1`

	rows, err := conn.Query(ctx, 
							query, 
							product.ID)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, fmt.Errorf("FAILED to query product by id: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		res_product, err := w.scanProductFromRows(rows)
		if err != nil {
			w.logger.Error().
					Ctx(ctx).
					Err(err).Send()
			return nil, err
		}
		return res_product, nil
	}

	w.logger.Warn().
			Ctx(ctx).
			Err(erro.ErrNotFound).Send()

	return nil, erro.ErrNotFound
}

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
					pending = pending + $4,
					reserved = reserved + $5,
					sold = sold + $6,
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
						inventory.Pending,
						inventory.Reserved,
						inventory.Sold,
					)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Str("func","UpdateInventory").
				Err(err).Send()
		return 0, fmt.Errorf("FAILED to update inventory: %w", err)
	}

	return row.RowsAffected(), nil
}