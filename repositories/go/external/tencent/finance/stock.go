package finance

import "fmt"

var (
	// stock type
	AStock                 StockType = "aStock"
	listStocksDefaultCount           = 200
	MarketType                       = map[string]string{
		"GP-A":     "主板",
		"GP-A-KCB": "科创板",
		"GP-A-CYB": "创业板",
	}
)

func ListStocks(options ...ListStocksOption) ([]Stock, error) {
	cfg := &listStocksConfig{
		typ:   AStock,
		count: listStocksDefaultCount,
	}
	for _, option := range options {
		option(cfg)
	}
	stocks := []Stock{}
	offset := 0
	count := 200
	for {
		data, err := getBoardRankList(cfg.typ.String(), WithOffset(offset), WithCount(count))
		if err != nil {
			return nil, fmt.Errorf("list stocks %v error: [%w]", cfg, err)
		}
		for _, node := range data.RankList {
			stocks = append(stocks, Stock{
				Code:   node.Code,
				Name:   node.Name,
				Market: MarketType[node.StockType],
			})
		}
		if len(stocks) >= cfg.count {
			break
		}
		offset += count
	}

	return stocks[:cfg.count], nil
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
