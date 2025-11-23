# go-inventory

    workload for POC purpose

# tables

    CREATE TABLE public.product (
        id 			BIGSERIAL 	NOT NULL,
        sku 		VARCHAR(100)	NOT NULL,
        type 		VARCHAR(100) NOT NULL,
        name 		VARCHAR(100) NOT NULL,
        status		VARCHAR(100) NOT NULL,
        created_at	timestamptz NOT NULL,
        updated_at	timestamptz NULL,
        CONSTRAINT 	product_pkey PRIMARY KEY (id)
    );

    CREATE UNIQUE INDEX product_sku_unique_idx ON public.product USING btree (sku);
    
    CREATE TABLE public.inventory (
        id 				BIGSERIAL	NOT NULL,
        fk_product_id	BIGSERIAL	NOT NULL,
        available		INT 		NOT null DEFAULT 0,
        reserved	 	INT 		NOT NULL DEFAULT 0,
        sold		 	INT 		NOT null DEFAULT 0,
        created_at 		timestamptz 	NOT NULL,
        updated_at 		timestamptz 	NULL,   
        CONSTRAINT inventory_pkey PRIMARY KEY (id)
    );

    ALTER TABLE public.inventory ADD CONSTRAINT inventory_fk_product_id_fkey 
    FOREIGN KEY (fk_product_id) REFERENCES public.product(id);
