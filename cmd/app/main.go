package main

import (
	"crawler/internal/config"
	wb "crawler/internal/parsers"
	"encoding/json"
	"log"
	"os"
)

func main() {
	parser := wb.NewWBParser()

	var c config.Config
	err := c.ReadEnv()

	if err != nil {
		log.Fatal(err.Error())
	}

	product, err := parser.GetProductInfo(&c)

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
