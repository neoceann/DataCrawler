package wb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"crawler/internal/config"
)

const (
	BaseURLProduct = "https://www.wildberries.ru/__internal/u-card/cards/v4/detail"
	BaseURLSearch  = "https://www.wildberries.ru/__internal/u-search/exactmatch/ru/common/v18/search"

	CommonParams = "appType=1&curr=rub&dest=-2133462&spp=30&hide_vflags=4294967296&ab_testing=false&lang=ru&locale=ru"
	SearchParams = "inheritFilters=false&resultset=catalog&suppressSpellcheck=false"

	ProductID = "200135094" //test
)

type WBResponse struct {
	Products []*WBProduct `json:"products"`
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
	URL := fmt.Sprintf("%s?%s&nm=%s",
		BaseURLProduct, CommonParams, ProductID)

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

	return resp.Products[0], nil

}

func (p *WBParser) GetTopProducts(cfg *config.Config, query, sortBy string, page, limit int) ([]*WBProduct, error) {
	q := url.QueryEscape(query)

	URL := fmt.Sprintf("%s?%s&%s&q1=%s&query=%s&sort=%s&page=%d&limit=%d",
		BaseURLSearch, CommonParams, SearchParams, q, q, sortBy, page, limit)

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

	return resp.Products, nil
}
