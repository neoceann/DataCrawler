package main

import (
	"context"
	"crawler/internal/app"
	"crawler/internal/shutdown"
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
		err := crawler.SaveProductsToDB(ctx, p)

		if err != nil {
			log.Fatal(err)
		}

		log.Println("\nProducts saved to DB and \"products.json\"")
	} else {
		log.Println("\nProducts saved to \"products.json\"")
	}
	crawler.SaveProductsToFile(p)
}
