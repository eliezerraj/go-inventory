package database

import (
		"context"
		"errors"
		"database/sql"

		"github.com/rs/zerolog"
		"github.com/jackc/pgx/v5"

		"github.com/go-inventory/shared/erro"
		"github.com/go-inventory/internal/domain/model"

		go_core_otel_trace "github.com/eliezerraj/go-core/v2/otel/trace"
		go_core_db_pg "github.com/eliezerraj/go-core/v2/database/postgre"
)

var tracerProvider go_core_otel_trace.TracerProvider

type WorkerRepository struct {
	DatabasePG *go_core_db_pg.DatabasePGServer
	logger		*zerolog.Logger
}

// Above new worker
func NewWorkerRepository(databasePG *go_core_db_pg.DatabasePGServer,
						appLogger *zerolog.Logger) *WorkerRepository{
	logger := appLogger.With().
						Str("package", "repo.database").
						Logger()
	logger.Info().
			Str("func","NewWorkerRepository").Send()

	return &WorkerRepository{
		DatabasePG: databasePG,
		logger: &logger,
	}
}

// Above get stats from database
func (w *WorkerRepository) Stat(ctx context.Context) (go_core_db_pg.PoolStats){
	w.logger.Info().
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
	// trace
	ctx, span := tracerProvider.SpanCtx(ctx, "database.AddProduct")
	defer span.End()

	w.logger.Info().
			Ctx(ctx).
			Str("func","AddProduct").Send()

	conn, err := w.DatabasePG.Acquire(ctx)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePG.Release(conn)

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
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, errors.New(err.Error())
	}

	// Set PK
	product.ID = id
	
	return product , nil
}

// About get a product
func (w *WorkerRepository) GetProduct(ctx context.Context, 
									product *model.Product) (*model.Product, error){
	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "database.GetProduct")
	defer span.End()

	w.logger.Info().
			Ctx(ctx).
			Str("func","GetProduct").Send()

	// db connection
	conn, err := w.DatabasePG.Acquire(ctx)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePG.Release(conn)

	// Prepare
	res_product := model.Product{}
	var nullUpdatedAt sql.NullTime

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
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

    if err := rows.Err(); err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Msg("fatal error closing rows")
        return nil, errors.New(err.Error())
    }

	for rows.Next() {
		err := rows.Scan(	&res_product.ID, 
							&res_product.Sku, 
							&res_product.Type,
							&res_product.Name,
							&res_product.Status,
							&res_product.CreatedAt,
							&nullUpdatedAt,
						)
		if err != nil {
			w.logger.Error().
					Ctx(ctx).
					Err(err).Send()
			return nil, errors.New(err.Error())
        }

		if nullUpdatedAt.Valid {
        	res_product.UpdatedAt = &nullUpdatedAt.Time
    	} else {
			res_product.UpdatedAt = nil
		}
		return &res_product, nil
	}

	w.logger.Warn().
			Ctx(ctx).
			Err(err).Send()

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
	ctx, span := tracerProvider.SpanCtx(ctx, "database.AddInventory")
	defer span.End()

	conn, err := w.DatabasePG.Acquire(ctx)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePG.Release(conn)

	//Prepare
	var id int

	// Query Execute
	query := `INSERT INTO inventory ( 	fk_product_id,
										available,
										reserved,
										sold,
										created_at) 
				VALUES($1, $2, $3, $4, $5) RETURNING id`

	row := tx.QueryRow(	ctx, 
						query,
						inventory.Product.ID,
						inventory.Available, 
						inventory.Reserved,
						inventory.Sold,
						inventory.CreatedAt)
						
	if err := row.Scan(&id); err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, errors.New(err.Error())
	}

	// Set PK
	inventory.ID = id
	
	return inventory , nil
}

// About get a Inventory
func (w *WorkerRepository) GetInventory(ctx context.Context, 
										inventory *model.Inventory) (*model.Inventory, error){
	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "database.GetInventory")
	defer span.End()

	w.logger.Info().
			Ctx(ctx).
			Str("func","GetInventory").Send()

	// db connection
	conn, err := w.DatabasePG.Acquire(ctx)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return nil, errors.New(err.Error())
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
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

    if err := rows.Err(); err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Msg("fatal error closing rows")
        return nil, errors.New(err.Error())
    }

	for rows.Next() {
		err := rows.Scan(	&res_product.ID, 
							&res_product.Sku, 
							&res_product.Type,
							&res_product.Name,
							&res_product.Status,
							&res_product.CreatedAt,
							&nullProductUpdatedAt,
							&res_inventory.ID, 
							&res_inventory.Available, 
							&res_inventory.Reserved, 
							&res_inventory.Sold,
							&res_inventory.CreatedAt,
							&nullInventoryUpdatedAt,
						)
		if err != nil {
			w.logger.Error().
					Ctx(ctx).
					Err(err).Send()
			return nil, errors.New(err.Error())
        }

		if nullProductUpdatedAt.Valid {
        	res_product.UpdatedAt = &nullProductUpdatedAt.Time
    	} else {
			res_product.UpdatedAt = nil
		}
		if nullInventoryUpdatedAt.Valid {
        	res_inventory.UpdatedAt = &nullInventoryUpdatedAt.Time
    	} else {
			res_inventory.UpdatedAt = nil
		}
		res_inventory.Product = res_product
		return &res_inventory, nil
	}

	w.logger.Warn().
			Ctx(ctx).
			Err(err).Send()

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
	ctx, span := tracerProvider.SpanCtx(ctx, "database.UpdateInventory")
	defer span.End()

	conn, err := w.DatabasePG.Acquire(ctx)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Err(err).Send()
		return 0, errors.New(err.Error())
	}
	defer w.DatabasePG.Release(conn)

	// Query Execute
	query := `UPDATE inventory
				SET available = available + $3,
					reserved = reserved + $4,
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
					)
	if err != nil {
		w.logger.Error().
				Ctx(ctx).
				Str("func","UpdateInventory").
				Err(err).Send()
		return 0, errors.New(err.Error())
	}

	return row.RowsAffected(), nil
}