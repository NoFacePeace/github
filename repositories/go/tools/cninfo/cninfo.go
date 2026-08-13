package tools

import (
	"context"
	"fmt"

	"github.com/NoFacePeace/github/repositories/go/external/cninfo"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetLatestReportToolInput struct {
	Stock string `json:"stock" jsonschema:"the stock code with market prefix, e.g. sz000001 or sh600547"`
}

type GetLatestReportToolOutput struct {
	Report *cninfo.Report `json:"report" jsonschema:"the latest periodic report title and document URL, or null when no report is found"`
}

type GetAnnualReportSummariesToolInput struct {
	Stock string `json:"stock" jsonschema:"the stock code with market prefix, e.g. sz000001 or sh600547"`
}

type GetAnnualReportSummariesToolOutput struct {
	Reports []cninfo.Report `json:"reports" jsonschema:"all annual report summary titles and document URLs for the stock"`
}

var GetLatestReportToolMeta = &mcp.Tool{
	Name:        "get_latest_report",
	Description: "get the title and document URL of the latest annual, semiannual, first-quarter, or third-quarter report for an A-share stock",
}

var GetAnnualReportSummariesToolMeta = &mcp.Tool{
	Name:        "get_annual_report_summaries",
	Description: "get the titles and document URLs of all annual report summaries for an A-share stock",
}

func GetLatestReportTool(ctx context.Context, req *mcp.CallToolRequest, input GetLatestReportToolInput) (
	*mcp.CallToolResult,
	GetLatestReportToolOutput,
	error,
) {
	report, err := cninfo.QueryLatestReport(ctx, input.Stock)
	if err != nil {
		return nil, GetLatestReportToolOutput{}, fmt.Errorf("get latest report: %w", err)
	}
	return nil, GetLatestReportToolOutput{Report: report}, nil
}

func GetAnnualReportSummariesTool(ctx context.Context, req *mcp.CallToolRequest, input GetAnnualReportSummariesToolInput) (
	*mcp.CallToolResult,
	GetAnnualReportSummariesToolOutput,
	error,
) {
	reports, err := cninfo.QueryAnnualReportSummaries(ctx, input.Stock)
	if err != nil {
		return nil, GetAnnualReportSummariesToolOutput{}, fmt.Errorf("get annual report summaries: %w", err)
	}
	return nil, GetAnnualReportSummariesToolOutput{Reports: reports}, nil
}
