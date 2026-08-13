// Package cninfo provides a client for CNINFO (巨潮资讯网) disclosure APIs.
package cninfo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultBaseURL    = "https://www.cninfo.com.cn"
	szseStockURL      = defaultBaseURL + "/new/data/szse_stock.json"
	szseStockFileName = "szse_stock.json"
)

const (
	annualReportSummaryPageSize = 30
	latestReportPageSize        = 30
	staticFileBaseURL           = "https://static.cninfo.com.cn/"
	categoryAnnualReport        = "category_ndbg_szsh"
	categorySemiAnnualReport    = "category_bndbg_szsh"
	categoryFirstQuarterReport  = "category_yjdbg_szsh"
	categoryThirdQuarterReport  = "category_sjdbg_szsh"
	financialReportCategories   = categoryAnnualReport + ";" + categorySemiAnnualReport + ";" + categoryFirstQuarterReport + ";" + categoryThirdQuarterReport
	annualReportSummaryKeyword  = "年度报告摘要"
)

var reportYearPattern = regexp.MustCompile(`(\d{4})年`)

// queryOption 用于配置公告查询的表单参数。
type queryOption func(url.Values)

// withStock 设置 stock，格式为“证券代码,组织机构 ID”；为空时查询全市场公告。
func withStock(stock string) queryOption {
	return func(form url.Values) {
		form.Set("stock", stock)
	}
}

// withTabName 设置 tabName，默认值 fulltext 表示全文检索。
func withTabName(tabName string) queryOption {
	return func(form url.Values) {
		form.Set("tabName", tabName)
	}
}

// withPageSize 设置 pageSize，即每页返回的公告数量。
func withPageSize(pageSize int) queryOption {
	return func(form url.Values) {
		form.Set("pageSize", strconv.Itoa(pageSize))
	}
}

// withPageNum 设置 pageNum，页码从 1 开始。
func withPageNum(pageNum int) queryOption {
	return func(form url.Values) {
		form.Set("pageNum", strconv.Itoa(pageNum))
	}
}

// withColumn 设置 column，即交易所标识，例如 szse 或 sse。
func withColumn(column string) queryOption {
	return func(form url.Values) {
		form.Set("column", column)
	}
}

// withCategory 设置 category，即以分号分隔的公告分类。
func withCategory(category string) queryOption {
	return func(form url.Values) {
		form.Set("category", category)
	}
}

// withPlate 设置 plate，即市场板块，例如 sz 或 sh。
func withPlate(plate string) queryOption {
	return func(form url.Values) {
		form.Set("plate", plate)
	}
}

// withDateRange 设置 seDate，即公告日期范围筛选条件。
func withDateRange(dateRange string) queryOption {
	return func(form url.Values) {
		form.Set("seDate", dateRange)
	}
}

// withSearchKey 设置 searchkey，即全文检索关键词。
func withSearchKey(searchKey string) queryOption {
	return func(form url.Values) {
		form.Set("searchkey", searchKey)
	}
}

// withSecID 设置 secid，即可选的证券内部标识。
func withSecID(secID string) queryOption {
	return func(form url.Values) {
		form.Set("secid", secID)
	}
}

// withSort 设置 sortName 和 sortType，即排序字段及方向。
func withSort(sortName, sortType string) queryOption {
	return func(form url.Values) {
		form.Set("sortName", sortName)
		form.Set("sortType", sortType)
	}
}

// withHLTitle 设置 isHLtitle；为 true 时，标题命中的关键词可能以 HTML 高亮标记返回。
func withHLTitle(highlight bool) queryOption {
	return func(form url.Values) {
		form.Set("isHLtitle", strconv.FormatBool(highlight))
	}
}

// Announcement is a disclosure announcement returned by CNINFO.
type Announcement struct {
	AnnouncementID      string `json:"announcementId"`
	AnnouncementTitle   string `json:"announcementTitle"`
	AnnouncementContent string `json:"announcementContent"`
	AnnouncementTime    int64  `json:"announcementTime"`
	SecCode             string `json:"secCode"`
	SecName             string `json:"secName"`
	OrgID               string `json:"orgId"`
	AdjunctURL          string `json:"adjunctUrl"`
	AdjunctSize         int64  `json:"adjunctSize"`
	AdjunctType         string `json:"adjunctType"`
	Important           bool   `json:"important"`
}

// Report 是报告的标题和文件地址。
type Report struct {
	Title string
	URL   string
}

// QueryResponse is the historical-announcements API response.
type QueryResponse struct {
	Announcements     []Announcement `json:"announcements"`
	TotalAnnouncement int            `json:"totalAnnouncement"`
	HasMore           bool           `json:"hasMore"`
}

// QueryAnnualReportSummaries 查询指定股票的全部年度报告摘要。
// stock 的格式为“市场前缀 + 证券代码”，例如“sz000001”或“sh600547”。
func QueryAnnualReportSummaries(ctx context.Context, stock string) ([]Report, error) {
	security, err := resolveSecurity(ctx, stock)
	if err != nil {
		return nil, err
	}
	return queryAnnualReportSummariesWithClient(ctx, http.DefaultClient, defaultBaseURL, security.stock())
}

// QueryLatestReport 查询指定股票最新的定期报告，包含年度、半年度、一季度和三季度报告。
// stock 的格式为“市场前缀 + 证券代码”，例如“sz000001”或“sh600547”。
// 未查询到公告时返回 nil, nil。
func QueryLatestReport(ctx context.Context, stock string) (*Report, error) {
	security, err := resolveSecurity(ctx, stock)
	if err != nil {
		return nil, err
	}
	return queryLatestReportWithClient(ctx, http.DefaultClient, defaultBaseURL, security.stock())
}

type security struct {
	code  string
	orgID string
}

func (s security) stock() string {
	return s.code + "," + s.orgID
}

type stockRecord struct {
	Code     string `json:"code"`
	Category string `json:"category"`
	OrgID    string `json:"orgId"`
}

type szseStockFile struct {
	StockList []stockRecord `json:"stockList"`
}

func resolveSecurity(ctx context.Context, stock string) (security, error) {
	stockFilePath, err := szseStockFilePath()
	if err != nil {
		return security{}, err
	}
	return resolveSecurityWithClient(ctx, http.DefaultClient, szseStockURL, stockFilePath, stock)
}

func resolveSecurityWithClient(ctx context.Context, httpClient *http.Client, stockURL, stockFilePath, stock string) (security, error) {
	_, code, err := splitStock(stock)
	if err != nil {
		return security{}, err
	}

	stockFile, err := readSzseStockFile(stockFilePath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if err := downloadSzseStockFile(ctx, httpClient, stockURL, stockFilePath); err != nil {
			return security{}, err
		}
		stockFile, err = readSzseStockFile(stockFilePath)
		if err != nil {
			return security{}, err
		}
	case err != nil:
		return security{}, err
	}

	if result, found := findStock(stockFile.StockList, code); found {
		return result, nil
	}

	if err := downloadSzseStockFile(ctx, httpClient, stockURL, stockFilePath); err != nil {
		return security{}, err
	}
	stockFile, err = readSzseStockFile(stockFilePath)
	if err != nil {
		return security{}, err
	}
	if result, found := findStock(stockFile.StockList, code); found {
		return result, nil
	}
	return security{}, fmt.Errorf("security not found: %s", stock)
}

func findStock(records []stockRecord, code string) (security, bool) {
	for _, result := range records {
		if result.Code == code && result.Category == "A股" && result.OrgID != "" {
			return security{code: result.Code, orgID: result.OrgID}, true
		}
	}
	return security{}, false
}

func szseStockFilePath() (string, error) {
	cacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("os.UserCacheDir: %w", err)
	}
	return filepath.Join(cacheDirectory, "cninfo", szseStockFileName), nil
}

func readSzseStockFile(stockFilePath string) (szseStockFile, error) {
	body, err := os.ReadFile(stockFilePath)
	if err != nil {
		return szseStockFile{}, fmt.Errorf("os.ReadFile: %w", err)
	}

	var stockFile szseStockFile
	if err := json.Unmarshal(body, &stockFile); err != nil {
		return szseStockFile{}, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return stockFile, nil
}

func downloadSzseStockFile(ctx context.Context, httpClient *http.Client, stockURL, stockFilePath string) error {
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, stockURL, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("http.Client.Do: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return fmt.Errorf("unexpected status %s: %s", response.Status, string(body))
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(stockFilePath), 0o755); err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(stockFilePath), szseStockFileName+"-*")
	if err != nil {
		return fmt.Errorf("os.CreateTemp: %w", err)
	}
	tempFilePath := tempFile.Name()
	defer os.Remove(tempFilePath)

	if _, err := tempFile.Write(body); err != nil {
		tempFile.Close()
		return fmt.Errorf("os.File.Write: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("os.File.Close: %w", err)
	}
	if err := os.Rename(tempFilePath, stockFilePath); err != nil {
		return fmt.Errorf("os.Rename: %w", err)
	}
	return nil
}

func splitStock(stock string) (market, code string, err error) {
	stock = strings.ToLower(strings.TrimSpace(stock))
	if len(stock) != 8 {
		return "", "", fmt.Errorf("invalid stock: %s", stock)
	}

	market, code = stock[:2], stock[2:]
	if market != "sz" && market != "sh" {
		return "", "", fmt.Errorf("invalid stock market: %s", market)
	}
	for _, char := range code {
		if char < '0' || char > '9' {
			return "", "", fmt.Errorf("invalid stock code: %s", code)
		}
	}
	return market, code, nil
}

// queryAnnouncements fetches historical announcements matching options.
func queryAnnouncements(ctx context.Context, options ...queryOption) (*QueryResponse, error) {
	return queryAnnouncementsWithClient(ctx, http.DefaultClient, defaultBaseURL, options...)
}

func queryLatestReportWithClient(ctx context.Context, httpClient *http.Client, baseURL, stock string) (*Report, error) {
	announcementsCount := 0
	var latestAnnouncement *Announcement
	latestPeriod := 0
	for pageNum := 1; ; pageNum++ {
		response, err := queryAnnouncementsWithClient(ctx, httpClient, baseURL,
			withStock(stock),
			withCategory(financialReportCategories),
			withPageSize(latestReportPageSize),
			withPageNum(pageNum),
		)
		if err != nil {
			return nil, err
		}

		for _, announcement := range response.Announcements {
			period, ok := reportPeriod(announcement.AnnouncementTitle)
			if !ok || isEnglishVersionReport(announcement.AnnouncementTitle) {
				continue
			}
			if latestAnnouncement == nil ||
				period > latestPeriod ||
				period == latestPeriod && isReportSummary(latestAnnouncement.AnnouncementTitle) && !isReportSummary(announcement.AnnouncementTitle) {
				announcement := announcement
				latestAnnouncement = &announcement
				latestPeriod = period
			}
		}
		announcementsCount += len(response.Announcements)
		if len(response.Announcements) == 0 ||
			response.TotalAnnouncement > 0 && announcementsCount >= response.TotalAnnouncement ||
			!response.HasMore && len(response.Announcements) < latestReportPageSize {
			break
		}
	}
	if latestAnnouncement == nil {
		return nil, nil
	}
	return &Report{
		Title: latestAnnouncement.AnnouncementTitle,
		URL:   staticFileBaseURL + latestAnnouncement.AdjunctURL,
	}, nil
}

func isEnglishVersionReport(title string) bool {
	title = strings.ToLower(title)
	return strings.Contains(title, "英文版") || strings.Contains(title, "english version")
}

func isReportSummary(title string) bool {
	return strings.Contains(title, "摘要")
}

func reportPeriod(title string) (int, bool) {
	matches := reportYearPattern.FindStringSubmatch(title)
	if len(matches) != 2 {
		return 0, false
	}
	year, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, false
	}

	quarter := 0
	switch {
	case strings.Contains(title, "第一季度报告"):
		quarter = 1
	case strings.Contains(title, "半年度报告"):
		quarter = 2
	case strings.Contains(title, "第三季度报告"):
		quarter = 3
	case strings.Contains(title, "年度报告"):
		quarter = 4
	default:
		return 0, false
	}
	return year*4 + quarter, true
}

func queryAnnualReportSummariesWithClient(ctx context.Context, httpClient *http.Client, baseURL, stock string) ([]Report, error) {
	summaries := make([]Report, 0)
	announcementsCount := 0
	for pageNum := 1; ; pageNum++ {
		response, err := queryAnnouncementsWithClient(ctx, httpClient, baseURL,
			withStock(stock),
			withCategory(categoryAnnualReport),
			withSearchKey(annualReportSummaryKeyword),
			withPageSize(annualReportSummaryPageSize),
			withPageNum(pageNum),
		)
		if err != nil {
			return nil, err
		}

		for _, announcement := range response.Announcements {
			summaries = append(summaries, Report{
				Title: announcement.AnnouncementTitle,
				URL:   staticFileBaseURL + announcement.AdjunctURL,
			})
		}
		announcementsCount += len(response.Announcements)
		if len(response.Announcements) == 0 ||
			response.TotalAnnouncement > 0 && announcementsCount >= response.TotalAnnouncement ||
			!response.HasMore && len(response.Announcements) < annualReportSummaryPageSize {
			break
		}
	}
	return summaries, nil
}

func queryAnnouncementsWithClient(ctx context.Context, httpClient *http.Client, baseURL string, options ...queryOption) (*QueryResponse, error) {
	form := url.Values{
		"tabName":   {"fulltext"},
		"pageSize":  {"30"},
		"pageNum":   {"1"},
		"isHLtitle": {"false"},
	}
	for _, option := range options {
		option(form)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/new/hisAnnouncement/query", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("http.Client.Do: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return nil, fmt.Errorf("unexpected status %s: %s", response.Status, string(body))
	}

	var result QueryResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("json.Decoder.Decode: %w", err)
	}
	return &result, nil
}
