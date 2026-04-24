package wb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"crawler/internal/config"
)

const (
	Dest           = "-2133462"
	BaseURLProduct = "https://www.wildberries.ru/__internal/u-card/cards/v4/detail"
	ProductID      = "200135094"
)

type WBResponse struct {
	Products []WBProduct `json:"products"`
}

type WBProduct struct {
	ID             int64   `json:"id"`
	Brand          string  `json:"brand"`
	Name           string  `json:"name"`
	Supplier       string  `json:"supplier"`
	SupplierRating float64 `json:"supplierRating"`
	ProductRating  float64 `json:"reviewRating"`
	FeedbackCount  int64   `json:"feedbacks"`
	Sizes          []struct {
		Price struct {
			BasicPrice  int64 `json:"basic"`
			ActualPrice int64 `json:"product"`
		} `json:"price"`
	} `json:"sizes"`
	TotalQuantity int64 `json:"totalQuantity"`
}

type WBParser struct {
	client  *http.Client
	baseURL string
}

func NewWBParser() *WBParser {
	return &WBParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *WBParser) GetProductInfo(cfg *config.Config) (*WBProduct, error) {
	URL := fmt.Sprintf("%s?appType=1&curr=rub&dest=%s&spp=30&hide_vflags=4294967296&ab_testing=false&lang=ru&nm=%s",
		BaseURLProduct, Dest, ProductID)

	req, _ := http.NewRequest("GET", URL, nil)
	req.Header.Set("deviceid", cfg.DeviceID)
	req.Header.Set("Cookie", fmt.Sprintf("x_wbaas_token=%s; _wbauid=%s; _wbauid=%s", cfg.XWbaasToken, cfg.Wbauid1, cfg.Wbauid2))
	req.Header.Set("user-agent", cfg.UserAgent)

	response, err := p.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("http %d: %s", response.StatusCode, string(body))
	}

	body, _ := io.ReadAll(response.Body)

	var resp WBResponse

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Products[0], nil

}
