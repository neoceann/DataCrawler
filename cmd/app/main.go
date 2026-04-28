package main

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/parser"
	"crawler/internal/repository/db"
	"encoding/json"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	var c config.Config
	var cdb config.DBConfig
	err := config.ReadEnv(&c, &cdb)

	if err != nil {
		log.Fatal(err.Error())
	}

	pool, err := db.NewPool(ctx, &cdb)

	if err != nil {
		log.Fatal("Failed to create pool:", err)
	}
	defer pool.Close()

	parser := parser.NewWBParser(db.New(pool.Pool))

	product, err := parser.GetProductInfo(&c)

	err = parser.SaveProductToDB(ctx, product)

	if err != nil {
		log.Fatal(err.Error())
	}

	products, err := parser.GetTopProducts(&c, "коврик для мышки", "popular", 1, 50)

	if err != nil {
		log.Fatal(err.Error())
	}

	output, _ := json.MarshalIndent(product, "", "  ")
	os.WriteFile("one_product.json", output, 0644)

	outputs, _ := json.MarshalIndent(products, "", "  ")
	os.WriteFile("products.json", outputs, 0644)
}
