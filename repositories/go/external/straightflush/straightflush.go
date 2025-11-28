package straightflush

import (
	"context"
	"fmt"

	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
	"github.com/NoFacePeace/github/repositories/go/utils/http"

	stdhttp "net/http"
)

type Block struct {
	Name string
}

func GetBlockList(ctx context.Context) ([]Block, error) {
	ret, err := getBlockList(ctx)
	if err != nil {
		return nil, fmt.Errorf("get block list error: [%w]", err)
	}
	if ret.StatusCode != 0 {
		return nil, fmt.Errorf("get block list status message: [%s]", ret.StatusMsg)
	}
	bs := []Block{}
	for _, v := range ret.Data.List {
		bs = append(bs, Block{
			Name: v.BlockName,
		})
	}
	return bs, nil
}

func getBlockList(ctx context.Context) (*getBlockListResp, error) {
	url := "https://dq.10jqka.com.cn/interval_calculation/block_info/v1/get_block_list"
	req := &getBlockListReq{}
	req.Type = 1
	req.HistoryInfo.HistoryType = "0"
	req.PageInfo.Page = 1
	req.PageInfo.PageSize = 100
	yesterday := datetime.Yesterday()
	req.HistoryInfo.StartDate = yesterday.Format(datetime.LayoutDate) + "093000"
	req.HistoryInfo.EndDate = yesterday.Format(datetime.LayoutDate) + "150000"
	resp := &getBlockListResp{}
	if err := http.Do(ctx, url, stdhttp.MethodPost, req, resp, http.WithContentTypeJson()); err != nil {
		return nil, fmt.Errorf("http do error: [%w]", err)
	}
	return resp, nil
}

type getBlockListReq struct {
	Type        int `json:"type"`
	HistoryInfo struct {
		HistoryType string `json:"history_type"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
	} `json:"history_info"`
	PageInfo struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
	} `json:"page_info"`
	SortInfo struct {
		SortField string `json:"sort_field"`
		SortType  string `json:"sort_type"`
	} `json:"sort_info"`
}

type getBlockListResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	Data       struct {
		Total int `json:"total"`
		List  []struct {
			Turnover             int64   `json:"turnover"`
			BlockMarket          string  `json:"block_market"`
			BlockCode            string  `json:"block_code"`
			BlockName            string  `json:"block_name"`
			MarginOfIncrease     float64 `json:"margin_of_increase"`
			NetInflowOfMainForce int64   `json:"net_inflow_of_main_force"`
			StockList            []struct {
				StockMarket      string  `json:"stock_market"`
				StockCode        string  `json:"stock_code"`
				StockName        string  `json:"stock_name"`
				MarginOfIncrease float64 `json:"margin_of_increase"`
			} `json:"stock_list"`
		} `json:"list"`
	} `json:"data"`
}
