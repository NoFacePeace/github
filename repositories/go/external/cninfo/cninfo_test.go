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
	if len(summaries) != 1 || summaries[0].Title != "2024年年度报告摘要" || summaries[0].ID != "finalpage/2025-03-15/1212345678.PDF" {
		t.Fatalf("summaries = %#v", summaries)
	}
}

func TestQueryLatestReport(t *testing.T) {
	requests := 0
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		requests++
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if got := form.Get("pageSize"); got != "30" {
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
			Body:       io.NopCloser(strings.NewReader(`{"announcements":[{"announcementTitle":"2025年年度报告摘要","adjunctUrl":"finalpage/2026-04-30/1225261225.PDF"},{"announcementTitle":"2026年第一季度报告","adjunctUrl":"finalpage/2026-04-30/1225261226.PDF"},{"announcementTitle":"2026年第一季度报告摘要","adjunctUrl":"finalpage/2026-04-30/1225261227.PDF"},{"announcementTitle":"2025年年度报告（英文版）","adjunctUrl":"finalpage/2026-06-17/1225374514.PDF"}],"totalAnnouncement":31,"hasMore":true}`)),
			Header:     make(http.Header),
		}, nil
	})}

	report, err := queryLatestReportWithClient(context.Background(), httpClient, "https://example.com", "000001,gssz0000001")
	if err != nil {
		t.Fatal(err)
	}
	if report == nil || report.Title != "2026年第一季度报告摘要" || report.ID != "finalpage/2026-04-30/1225261227.PDF" {
		t.Fatalf("report = %#v", report)
	}
	if requests != 1 {
		t.Errorf("requests = %d, want 1", requests)
	}
}

func TestQueryReportsDoesNotDuplicateLatestAnnualReport(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		switch form.Get("category") {
		case financialReportCategories:
			return jsonResponse(`{"announcements":[{"announcementTitle":"2025年年度报告摘要","adjunctUrl":"finalpage/2026-03-20/2025.PDF"}],"totalAnnouncement":1}`)
		case categoryAnnualReport:
			return jsonResponse(`{"announcements":[{"announcementTitle":"2025年年度报告摘要","adjunctUrl":"finalpage/2026-03-20/2025.PDF"},{"announcementTitle":"2024年年度报告摘要","adjunctUrl":"finalpage/2025-03-20/2024.PDF"}],"totalAnnouncement":2}`)
		default:
			t.Fatalf("unexpected category %q", form.Get("category"))
			return nil, nil
		}
	})}

	reports, err := queryReportsWithClient(context.Background(), httpClient, "https://example.com", "000001,gssz0000001")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 2 || reports[0].ID != "finalpage/2026-03-20/2025.PDF" || reports[1].ID != "finalpage/2025-03-20/2024.PDF" {
		t.Fatalf("reports = %#v", reports)
	}
}

func TestQueryReportsPrependsLatestNonAnnualReport(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		switch form.Get("category") {
		case financialReportCategories:
			return jsonResponse(`{"announcements":[{"announcementTitle":"2026年半年度报告摘要","adjunctUrl":"finalpage/2026-08-15/2026-half.PDF"}],"totalAnnouncement":1}`)
		case categoryAnnualReport:
			return jsonResponse(`{"announcements":[{"announcementTitle":"2025年年度报告摘要","adjunctUrl":"finalpage/2026-03-20/2025.PDF"}],"totalAnnouncement":1}`)
		default:
			t.Fatalf("unexpected category %q", form.Get("category"))
			return nil, nil
		}
	})}

	reports, err := queryReportsWithClient(context.Background(), httpClient, "https://example.com", "000001,gssz0000001")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 2 || reports[0].ID != "finalpage/2026-08-15/2026-half.PDF" || reports[1].ID != "finalpage/2026-03-20/2025.PDF" {
		t.Fatalf("reports = %#v", reports)
	}
}

func TestGetReport(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", request.Method)
		}
		if got, want := request.URL.String(), "https://example.com/finalpage/2026-08-15/1225475343.PDF"; got != want {
			t.Errorf("URL = %q, want %q", got, want)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader("%PDF-1.7")),
			Header:     make(http.Header),
		}, nil
	})}

	report, err := getReportWithClient(context.Background(), httpClient, "https://example.com", "finalpage/2026-08-15/1225475343.PDF")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(report), "%PDF-1.7"; got != want {
		t.Errorf("report = %q, want %q", got, want)
	}
}

func TestGetReportRejectsInvalidID(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		t.Fatal("unexpected request")
		return nil, nil
	})}

	_, err := getReportWithClient(context.Background(), httpClient, "https://example.com", "../secret.pdf")
	if err == nil {
		t.Fatal("getReportWithClient returned nil error")
	}
}

func TestGetReportWithCacheReturnsCachedReport(t *testing.T) {
	cacheFilePath := filepath.Join(t.TempDir(), "report.pdf")
	if err := os.WriteFile(cacheFilePath, []byte("%PDF-cached"), 0o600); err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		t.Fatal("unexpected download")
		return nil, nil
	})}

	reportPath, err := getReportWithCache(context.Background(), httpClient, "https://example.com", cacheFilePath, "finalpage/2026-08-15/1225475343.PDF")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := reportPath, cacheFilePath; got != want {
		t.Errorf("report path = %q, want %q", got, want)
	}
}

func TestGetReportWithCacheDownloadsMissingReport(t *testing.T) {
	cacheFilePath := filepath.Join(t.TempDir(), "report.pdf")
	httpClient := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader("%PDF-downloaded")),
			Header:     make(http.Header),
		}, nil
	})}

	reportPath, err := getReportWithCache(context.Background(), httpClient, "https://example.com", cacheFilePath, "finalpage/2026-08-15/1225475343.PDF")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := reportPath, cacheFilePath; got != want {
		t.Errorf("report path = %q, want %q", got, want)
	}
	cachedReport, err := os.ReadFile(cacheFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(cachedReport), "%PDF-downloaded"; got != want {
		t.Errorf("cached report = %q, want %q", got, want)
	}
}

func TestIsEnglishVersionReport(t *testing.T) {
	tests := []struct {
		title string
		want  bool
	}{
		{title: "2025年年度报告（英文版）", want: true},
		{title: "2025 Annual Report (English Version)", want: true},
		{title: "2026年一季度报告", want: false},
	}
	for _, test := range tests {
		if got := isEnglishVersionReport(test.title); got != test.want {
			t.Errorf("isEnglishVersionReport(%q) = %t", test.title, got)
		}
	}
}

func TestReportPeriod(t *testing.T) {
	tests := []struct {
		title string
		want  int
	}{
		{title: "中国船舶工业股份有限公司2025年年度报告摘要", want: 2025*4 + 4},
		{title: "中国船舶工业股份有限公司2026年第一季度报告", want: 2026*4 + 1},
		{title: "中国船舶2025年半年度报告", want: 2025*4 + 2},
	}
	for _, test := range tests {
		got, ok := reportPeriod(test.title)
		if !ok || got != test.want {
			t.Errorf("reportPeriod(%q) = (%d, %t)", test.title, got, ok)
		}
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

func jsonResponse(body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}, nil
}
