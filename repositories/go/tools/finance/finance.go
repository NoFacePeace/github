package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/NoFacePeace/github/repositories/go/external/tencent/finance"
	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListStocksToolInput struct {
	Count int `json:"count,omitempty" jsonschema:"the number of stocks to return, defaults to all when omitted or non-positive"`
}

type ListStocksToolOutput struct {
	Stocks []finance.Stock `json:"stocks" jsonschema:"the list of stocks"`
}

var ListStocksToolMeta = &mcp.Tool{
	Name:        "list_stocks",
	Description: "list A-share stocks with code, name and market; count <= 0 returns all",
}

func ListStocksTool(ctx context.Context, req *mcp.CallToolRequest, input ListStocksToolInput) (
	*mcp.CallToolResult,
	ListStocksToolOutput,
	error,
) {
	options := []finance.ListStocksOption{}
	if input.Count > 0 {
		options = append(options, finance.WithListStocksCount(input.Count))
	}
	stocks, err := finance.ListStocks(options...)
	if err != nil {
		return nil, ListStocksToolOutput{}, fmt.Errorf("list stocks: [%w]", err)
	}
	return nil, ListStocksToolOutput{Stocks: stocks}, nil
}

type GetKlineSinceToolInput struct {
	Code     string `json:"code" jsonschema:"the stock code, e.g. sh600941"`
	FromDate string `json:"fromDate" jsonschema:"the start date inclusive, format 2006-01-02"`
}

type GetKlineSinceToolOutput struct {
	Points []finance.Point `json:"points" jsonschema:"the kline points from fromDate to today"`
}

var GetKlineSinceToolMeta = &mcp.Tool{
	Name:        "get_kline_since",
	Description: "get kline points from a given date (inclusive) to today for a stock",
}

func GetKlineSinceTool(ctx context.Context, req *mcp.CallToolRequest, input GetKlineSinceToolInput) (
	*mcp.CallToolResult,
	GetKlineSinceToolOutput,
	error,
) {
	fromDate, err := time.Parse(datetime.LayoutDateWithDash, input.FromDate)
	if err != nil {
		return nil, GetKlineSinceToolOutput{}, fmt.Errorf("get kline since parse fromDate %s: [%w]", input.FromDate, err)
	}
	points, err := finance.GetKlineSince(input.Code, fromDate)
	if err != nil {
		return nil, GetKlineSinceToolOutput{}, fmt.Errorf("get kline since: [%w]", err)
	}
	return nil, GetKlineSinceToolOutput{Points: points}, nil
}
