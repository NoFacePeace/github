package finance

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	// stock type
	AStock           StockType = "aStock"
	listStocksPageSize          = 200
	MarketType                  = map[string]string{
		"GP-A":     "主板",
		"GP-A-KCB": "科创板",
		"GP-A-CYB": "创业板",
	}
)

func ListStocks(options ...ListStocksOption) ([]Stock, error) {
	cfg := &listStocksConfig{
		typ: AStock,
	}
	for _, option := range options {
		option(cfg)
	}
	bounded := cfg.count > 0
	stocks := []Stock{}
	offset := 0
	for {
		page := listStocksPageSize
		if bounded {
			remaining := cfg.count - len(stocks)
			if remaining <= 0 {
				break
			}
			page = min(remaining, listStocksPageSize)
		}
		data, err := getBoardRankList(cfg.typ.String(), WithOffset(offset), WithCount(page))
		if err != nil {
			return nil, fmt.Errorf("get board rank list: [%w]", err)
		}
		if data == nil {
			break
		}
		for _, node := range data.RankList {
			stocks = append(stocks, Stock{
				Code:       node.Code,
				Name:       node.Name,
				Market:     MarketType[node.StockType],
				TotalValue: node.Zsz,
				FlowValue:  node.Ltsz,
			})
		}
		if (bounded && len(stocks) >= cfg.count) || len(data.RankList) < page || data.Total > 0 && len(stocks) >= data.Total {
			break
		}
		offset += page
	}

	if bounded {
		return stocks[:min(len(stocks), cfg.count)], nil
	}
	return stocks, nil
}

type StockType string

func (typ StockType) String() string {
	return string(typ)
}

type listStocksConfig struct {
	typ   StockType
	count int
}

type ListStocksOption func(*listStocksConfig)

func WithStockType(typ StockType) ListStocksOption {
	return func(cfg *listStocksConfig) {
		cfg.typ = typ
	}
}

func WithListStocksCount(count int) ListStocksOption {
	return func(cfg *listStocksConfig) {
		cfg.count = count
	}
}

// ListStocksToolInput is the input schema for the ListStocks MCP tool.
type ListStocksToolInput struct {
	// Count is the number of stocks to return. Defaults to all when omitted or non-positive.
	Count int `json:"count,omitempty" jsonschema:"the number of stocks to return, defaults to all when omitted or non-positive"`
}

// ListStocksToolOutput is the output schema for the ListStocks MCP tool.
type ListStocksToolOutput struct {
	Stocks []Stock `json:"stocks" jsonschema:"the list of stocks"`
}

// ListStocksToolMeta is the MCP tool metadata for ListStocksTool.
var ListStocksToolMeta = &mcp.Tool{
	Name:        "list_stocks",
	Description: "list A-share stocks with code, name and market; count <= 0 returns all",
}

// ListStocksTool wraps ListStocks as an MCP tool.
func ListStocksTool(ctx context.Context, req *mcp.CallToolRequest, input ListStocksToolInput) (
	*mcp.CallToolResult,
	ListStocksToolOutput,
	error,
) {
	options := []ListStocksOption{}
	if input.Count > 0 {
		options = append(options, WithListStocksCount(input.Count))
	}
	stocks, err := ListStocks(options...)
	if err != nil {
		return nil, ListStocksToolOutput{}, fmt.Errorf("list stocks: [%w]", err)
	}
	return nil, ListStocksToolOutput{Stocks: stocks}, nil
}
