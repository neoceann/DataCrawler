package wb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"crawler/internal/config"
	"crawler/internal/parser"
	"crawler/internal/repository/db"
)

const (
	BaseURLProduct = "https://www.wildberries.ru/__internal/u-card/cards/v4/detail"
	BaseURLSearch  = "https://www.wildberries.ru/__internal/u-search/exactmatch/ru/common/v18/search"

	CommonParams = "appType=1&curr=rub&dest=-2133462&spp=30&hide_vflags=4294967296&ab_testing=false&lang=ru&locale=ru"
	SearchParams = "inheritFilters=false&resultset=catalog&suppressSpellcheck=false"
)

type WBParser struct {
	client  *http.Client
	queries *db.Queries
	cfg     *config.Config
}

func NewWBParser(q *db.Queries, cfg *config.Config) *WBParser {
	return &WBParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		queries: q,
		cfg:     cfg,
	}
}

func (p *WBParser) Name() string {
	return config.WB
}

func (p *WBParser) SaveProductToDB(ctx context.Context, product *parser.BaseProduct) error {
	return p.queries.CreateProduct(ctx, *product.ToDbParams())
}

func (p *WBParser) GetProductByID(ctx context.Context, productID string) (*parser.BaseProduct, error) {
	URL := fmt.Sprintf("%s?%s&nm=%s",
		BaseURLProduct, CommonParams, productID)

	req, _ := http.NewRequestWithContext(ctx, "GET", URL, nil)
	req.Header.Set("deviceid", p.cfg.WBDeviceID)
	req.Header.Set("Cookie", fmt.Sprintf("x_wbaas_token=%s; _wbauid=%s; _wbauid=%s", p.cfg.WBXWbaasToken, p.cfg.WBWbauid1, p.cfg.WBWbauid2))
	req.Header.Set("user-agent", p.cfg.WBUserAgent)

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

	var resp WBTopProducts

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return resp.Products[0].ToBaseProduct(), nil

}

func (p *WBParser) GetTopProducts(ctx context.Context, s *config.SearchConfig) ([]*parser.BaseProduct, error) {
	q := url.QueryEscape(s.Query)

	URL := fmt.Sprintf("%s?%s&%s&q1=%s&query=%s&sort=%s&page=%d&limit=%d",
		BaseURLSearch, CommonParams, SearchParams, q, q, s.SortBy, s.Page, s.Limit)

	req, _ := http.NewRequestWithContext(ctx, "GET", URL, nil)
	req.Header.Set("deviceid", p.cfg.WBDeviceID)
	req.Header.Set("Cookie", fmt.Sprintf("x_wbaas_token=%s; _wbauid=%s; _wbauid=%s", p.cfg.WBXWbaasToken, p.cfg.WBWbauid1, p.cfg.WBWbauid2))
	req.Header.Set("user-agent", p.cfg.WBUserAgent)

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

	var resp WBTopProducts

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return resp.ToBaseProducts(), nil
}
