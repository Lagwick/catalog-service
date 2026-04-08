    CREATE TABLE IF NOT EXISTS product (
        id BIGSERIAL NOT NULL UNIQUE,
        guid UUID NOT NULL PRIMARY KEY,
        name VARCHAR(255) NOT NULL UNIQUE,
        description TEXT,
        price DECIMAL(12,2) NOT NULL,
        category_guid UUID NOT NULL REFERENCES category(guid) ON DELETE RESTRICT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );