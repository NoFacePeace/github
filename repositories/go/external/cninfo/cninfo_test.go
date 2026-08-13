package cninfo

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQueryAnnouncements(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", request.Method)
		}
		if request.Header.Get("Content-Type") != "application/x-www-form-urlencoded; charset=UTF-8" {
			t.Fatalf("Content-Type = %q", request.Header.Get("Content-Type"))
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if got := form.Get("stock"); got != "000001,gssz0000001" {
			t.Errorf("stock = %q", got)
		}
		if got := form.Get("pageSize"); got != "30" {
			t.Errorf("pageSize = %q", got)
		}
		if got := form.Get("isHLtitle"); got != "true" {
			t.Errorf("isHLtitle = %q", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"announcements":[{"announcementId":"1","announcementTitle":"annual report","secCode":"000001"}],"totalAnnouncement":1,"hasMore":false}`)),
			Header:     make(http.Header),
		}, nil
	})}
	response, err := queryAnnouncementsWithClient(
		context.Background(), httpClient, "https://example.com",
		withStock("000001,gssz0000001"),
		withHLTitle(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.TotalAnnouncement != 1 || len(response.Announcements) != 1 {
		t.Fatalf("response = %#v", response)
	}
}

func TestQueryAnnualReportSummaries(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if got := form.Get("category"); got != categoryAnnualReport {
			t.Errorf("category = %q", got)
		}
		if got := form.Get("searchkey"); got != annualReportSummaryKeyword {
			t.Errorf("searchkey = %q", got)
		}
		if got := form.Get("stock"); got != "000001,gssz0000001" {
			t.Errorf("stock = %q", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"announcements":[{"announcementId":"2","announcementTitle":"2024年年度报告摘要","adjunctUrl":"finalpage/2025-03-15/1212345678.PDF"}],"totalAnnouncement":1}`)),
			Header:     make(http.Header),
		}, nil
	})}

	summaries, err := queryAnnualReportSummariesWithClient(context.Background(), httpClient, "https://example.com", "000001,gssz0000001")
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 || summaries[0].Title != "2024年年度报告摘要" || summaries[0].URL != "https://static.cninfo.com.cn/finalpage/2025-03-15/1212345678.PDF" {
		t.Fatalf("summaries = %#v", summaries)
	}
}

func TestQueryLatestReport(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if got := form.Get("pageSize"); got != "1" {
			t.Errorf("pageSize = %q", got)
		}
		if got := form.Get("category"); got != financialReportCategories {
			t.Errorf("category = %q", got)
		}
		if got := form.Get("searchkey"); got != "" {
			t.Errorf("searchkey = %q", got)
		}
		if got := form.Get("stock"); got != "000001,gssz0000001" {
			t.Errorf("stock = %q", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"announcements":[{"announcementTitle":"2025年第三季度报告","adjunctUrl":"finalpage/2026-10-30/1225022886.PDF"}],"totalAnnouncement":26}`)),
			Header:     make(http.Header),
		}, nil
	})}

	report, err := queryLatestReportWithClient(context.Background(), httpClient, "https://example.com", "000001,gssz0000001")
	if err != nil {
		t.Fatal(err)
	}
	if report == nil || report.Title != "2025年第三季度报告" || report.URL != "https://static.cninfo.com.cn/finalpage/2026-10-30/1225022886.PDF" {
		t.Fatalf("report = %#v", report)
	}
}

func TestResolveSecurity(t *testing.T) {
	stockFilePath := filepath.Join(t.TempDir(), szseStockFileName)
	if err := os.WriteFile(stockFilePath, []byte(`{"stockList":[{"code":"000001","category":"A股","orgId":"gssz0000001"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		t.Fatal("unexpected download")
		return nil, nil
	})}

	security, err := resolveSecurityWithClient(context.Background(), httpClient, "https://example.com/szse_stock.json", stockFilePath, "sz000001")
	if err != nil {
		t.Fatal(err)
	}
	if got := security.stock(); got != "000001,gssz0000001" {
		t.Errorf("stock = %q", got)
	}
}

func TestResolveSecuritySupportsShanghaiStock(t *testing.T) {
	stockFilePath := filepath.Join(t.TempDir(), szseStockFileName)
	if err := os.WriteFile(stockFilePath, []byte(`{"stockList":[{"code":"600547","category":"A股","orgId":"gssh0600547"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		t.Fatal("unexpected download")
		return nil, nil
	})}

	security, err := resolveSecurityWithClient(context.Background(), httpClient, "https://example.com/szse_stock.json", stockFilePath, "sh600547")
	if err != nil {
		t.Fatal(err)
	}
	if got := security.stock(); got != "600547,gssh0600547" {
		t.Errorf("stock = %q", got)
	}
}

func TestResolveSecurityDownloadsMissingFile(t *testing.T) {
	stockFilePath := filepath.Join(t.TempDir(), szseStockFileName)
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet {
			t.Errorf("method = %s", request.Method)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"stockList":[{"code":"002594","category":"A股","orgId":"gshk0001211"}]}`)),
			Header:     make(http.Header),
		}, nil
	})}

	security, err := resolveSecurityWithClient(context.Background(), httpClient, "https://example.com/szse_stock.json", stockFilePath, "sz002594")
	if err != nil {
		t.Fatal(err)
	}
	if got := security.stock(); got != "002594,gshk0001211" {
		t.Errorf("stock = %q", got)
	}
	if _, err := os.Stat(stockFilePath); err != nil {
		t.Fatal(err)
	}
}

func TestResolveSecurityRedownloadsWhenNotFound(t *testing.T) {
	stockFilePath := filepath.Join(t.TempDir(), szseStockFileName)
	if err := os.WriteFile(stockFilePath, []byte(`{"stockList":[{"code":"000001","category":"A股","orgId":"gssz0000001"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	downloads := 0
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		downloads++
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"stockList":[{"code":"002594","category":"A股","orgId":"gshk0001211"}]}`)),
			Header:     make(http.Header),
		}, nil
	})}

	security, err := resolveSecurityWithClient(context.Background(), httpClient, "https://example.com/szse_stock.json", stockFilePath, "sz002594")
	if err != nil {
		t.Fatal(err)
	}
	if downloads != 1 {
		t.Errorf("downloads = %d", downloads)
	}
	if got := security.stock(); got != "002594,gshk0001211" {
		t.Errorf("stock = %q", got)
	}
}

func TestSplitStock(t *testing.T) {
	tests := []struct {
		stock      string
		wantMarket string
		wantCode   string
		wantErr    bool
	}{
		{stock: "SZ000001", wantMarket: "sz", wantCode: "000001"},
		{stock: "sh600547", wantMarket: "sh", wantCode: "600547"},
		{stock: "000001", wantErr: true},
	}
	for _, test := range tests {
		market, code, err := splitStock(test.stock)
		if (err != nil) != test.wantErr {
			t.Errorf("splitStock(%q) error = %v", test.stock, err)
		}
		if market != test.wantMarket || code != test.wantCode {
			t.Errorf("splitStock(%q) = (%q, %q)", test.stock, market, code)
		}
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (fn roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}
