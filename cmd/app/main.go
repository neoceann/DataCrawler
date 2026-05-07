package main

import (
	"context"
	"crawler/internal/app"
	"crawler/internal/shutdown"
	"encoding/json"
	"os"
	"log"
)

func main() {
	ctx, cancel := shutdown.GracefulShutdown((context.Background()))
	defer cancel()

	crawler, err := app.New(ctx)

	if err != nil {
		log.Fatal("Failed to run app:%w", err)
	}
	defer crawler.Close()

	p, err := crawler.GetTopProducts(ctx)

	if err != nil {
		log.Fatal(err.Error())
	}

	if crawler.Config.SaveToDB {
		crawler.SaveProductsToDB(ctx, p)
	} else {
		outputs, _ := json.MarshalIndent(p, "", "  ")
		os.WriteFile("products.json", outputs, 0644)
	}
}
