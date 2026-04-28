package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"crawler/internal/config"
	"crawler/internal/repository/db"
	"crawler/internal/repository/mapper"
)

const (
	BaseURLProduct = "https://www.wildberries.ru/__internal/u-card/cards/v4/detail"
	BaseURLSearch  = "https://www.wildberries.ru/__internal/u-search/exactmatch/ru/common/v18/search"

	CommonParams = "appType=1&curr=rub&dest=-2133462&spp=30&hide_vflags=4294967296&ab_testing=false&lang=ru&locale=ru"
	SearchParams = "inheritFilters=false&resultset=catalog&suppressSpellcheck=false"

	ProductID = "200135094" //test
)

type WBParser struct {
	client  *http.Client
	queries *db.Queries
}

func NewWBParser(q *db.Queries) *WBParser {
	return &WBParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		queries: q,
	}
}

func (p *WBParser) SaveProductToDB(ctx context.Context, wbp *mapper.WBProduct) error {
	return p.queries.CreateProduct(ctx, *wbp.ToDbParams())
}

func (p *WBParser) GetProductInfo(cfg *config.Config) (*mapper.WBProduct, error) {
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

	var resp mapper.WBResponse

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return resp.Products[0], nil

}

func (p *WBParser) GetTopProducts(cfg *config.Config, query, sortBy string, page, limit int) ([]*mapper.WBProduct, error) {
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

	var resp mapper.WBResponse

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return resp.Products, nil
}
