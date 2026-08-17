package tools

import (
	"context"
	"fmt"

	"github.com/NoFacePeace/github/repositories/go/external/cninfo"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type QueryReportsToolInput struct {
	Stock string `json:"stock" jsonschema:"the stock code with market prefix, e.g. sz000001 or sh600547"`
}

type QueryReportsToolOutput struct {
	Reports []cninfo.Report `json:"reports" jsonschema:"the latest report followed by annual report summaries"`
}

type QueryReportsByDateToolInput struct {
	Date string `json:"date,omitempty" jsonschema:"the disclosure date in YYYY-MM-DD format; defaults to today when omitted"`
}

type QueryReportsByDateToolOutput struct {
	Reports []cninfo.Report `json:"reports" jsonschema:"all periodic reports disclosed today"`
}

type GetReportToolInput struct {
	ReportID string `json:"reportId" jsonschema:"the CNINFO report ID, e.g. finalpage/2026-08-15/1225475343.PDF"`
}

type GetReportToolOutput struct {
	ReportID string `json:"reportId" jsonschema:"the requested CNINFO report ID"`
	Path     string `json:"path" jsonschema:"the local PDF path"`
}

var QueryReportsToolMeta = &mcp.Tool{
	Name:        "query_reports",
	Description: "get the latest periodic report and annual report summaries for an A-share stock",
}

var QueryReportsByDateToolMeta = &mcp.Tool{
	Name:        "query_reports_by_date",
	Description: "get all periodic reports disclosed on a date; defaults to today",
}

var GetReportToolMeta = &mcp.Tool{
	Name:        "get_report",
	Description: "download a CNINFO report PDF to the local cache and return its path",
}

func QueryReportsTool(ctx context.Context, req *mcp.CallToolRequest, input QueryReportsToolInput) (
	*mcp.CallToolResult,
	QueryReportsToolOutput,
	error,
) {
	reports, err := cninfo.QueryReports(ctx, input.Stock)
	if err != nil {
		return nil, QueryReportsToolOutput{}, fmt.Errorf("query reports: %w", err)
	}
	return nil, QueryReportsToolOutput{Reports: reports}, nil
}

func QueryReportsByDateTool(ctx context.Context, req *mcp.CallToolRequest, input QueryReportsByDateToolInput) (
	*mcp.CallToolResult,
	QueryReportsByDateToolOutput,
	error,
) {
	reports, err := cninfo.QueryReportsByDate(ctx, input.Date)
	if err != nil {
		return nil, QueryReportsByDateToolOutput{}, fmt.Errorf("query reports by date: %w", err)
	}
	return nil, QueryReportsByDateToolOutput{Reports: reports}, nil
}

func GetReportTool(ctx context.Context, req *mcp.CallToolRequest, input GetReportToolInput) (
	*mcp.CallToolResult,
	GetReportToolOutput,
	error,
) {
	reportPath, err := cninfo.GetReport(ctx, input.ReportID)
	if err != nil {
		return nil, GetReportToolOutput{}, fmt.Errorf("get report: %w", err)
	}
	return nil, GetReportToolOutput{
		ReportID: input.ReportID,
		Path:     reportPath,
	}, nil
}
