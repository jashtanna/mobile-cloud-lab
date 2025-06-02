CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    image VARCHAR(255),
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    qty INTEGER NOT NULL CHECK (qty >= 0),
    out_of_stock BOOLEAN GENERATED ALWAYS AS (qty = 0) STORED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_out_of_stock ON products(out_of_stock);