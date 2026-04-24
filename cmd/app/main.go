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

	output, _ := json.MarshalIndent(product, "", "  ")
	os.WriteFile("products.json", output, 0644)
}
