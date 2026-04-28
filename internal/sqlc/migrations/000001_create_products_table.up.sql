CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    marketplace VARCHAR(20) NOT NULL,
    product_id VARCHAR(200) NOT NULL,
    brand VARCHAR(200),
    name TEXT NOT NULL,
    supplier VARCHAR(200) NOT NULL,
    supplier_rating DOUBLE PRECISION NOT NULL,
    product_rating DOUBLE PRECISION NOT NULL,
    feedbacks BIGINT NOT NULL,
    price_basic BIGINT NOT NULL,
    price_actual BIGINT NOT NULL,
    quantity BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(marketplace, product_id)
);

CREATE INDEX IF NOT EXISTS idx_products_product_id ON products(product_id);
CREATE INDEX IF NOT EXISTS idx_products_actual_price ON products(price_actual);
CREATE INDEX IF NOT EXISTS idx_products_product_rating ON products(product_rating);
CREATE INDEX IF NOT EXISTS idx_products_updated_at ON products(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_marketplace ON products(marketplace);