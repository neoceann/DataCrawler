--name CreateWBProduct :exec
INSERT INTO wb_products (
    marketplace,
    product_id,
    brand,
    name,
    supplier,
    supplier_rating,
    product_rating,
    feedbacks,
    price_basic,
    price_actual,
    quantity
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (marketplace, product_id) 
DO UPDATE SET 
    (brand, name, supplier, supplier_rating, product_rating, feedbacks, price_basic, price_actual, quantity, updated_at) = 
    (EXCLUDED.brand, EXCLUDED.name, EXCLUDED.supplier, EXCLUDED.supplier_rating, 
     EXCLUDED.product_rating, EXCLUDED.feedbacks, EXCLUDED.price_basic, EXCLUDED.price_actual, EXCLUDED.quantity, NOW())
WHERE products.name IS DISTINCT FROM EXCLUDED.name
   OR products.brand IS DISTINCT FROM EXCLUDED.brand
   OR products.price IS DISTINCT FROM EXCLUDED.price;