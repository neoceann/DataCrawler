package main

import (
	"context"
	"crawler/internal/app"
	"crawler/internal/config"
	"crawler/internal/shutdown"
	"encoding/json"
	"log"
	"os"
)

func main() {
	ctx, cancel := shutdown.GracefulShutdown((context.Background()))
	defer cancel()

	crawler, err := app.New(ctx)

	if err != nil {
		log.Fatal("Failed to run app:%w", err)
	}
	defer crawler.Close()

	product, err := crawler.GetProductByID(ctx, config.WB, "200135094")

	if err != nil {
		log.Fatal(err.Error())
	}

	output, _ := json.MarshalIndent(product, "", "  ")
	os.WriteFile("one_product.json", output, 0644)

	// outputs, _ := json.MarshalIndent(products, "", "  ")
	// os.WriteFile("products.json", outputs, 0644)
}
