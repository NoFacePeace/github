package finance

import "fmt"

var (
	// stock type
	AStock                 StockType = "aStock"
	listStocksDefaultCount           = 200
)

func ListStocks(options ...ListStocksOption) ([]Stock, error) {
	cfg := &listStocksConfig{
		typ:   AStock,
		count: listStocksDefaultCount,
	}
	for _, option := range options {
		option.apply(cfg)
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
				Code: node.Code,
				Name: node.Name,
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

func (typ StockType) apply(cfg *listStocksConfig) {
	cfg.typ = typ
}

type listStocksConfig struct {
	typ   StockType
	count int
}

type ListStocksOption interface {
	apply(*listStocksConfig)
}

func WithStockType(typ StockType) ListStocksOption {
	return AStock
}

func WithListStocksCount(count int) ListStocksOption {
	return &listStocksCountOption{
		count: count,
	}
}

type listStocksCountOption struct {
	count int
}

func (o *listStocksCountOption) apply(cfg *listStocksConfig) {
	cfg.count = o.count
}
