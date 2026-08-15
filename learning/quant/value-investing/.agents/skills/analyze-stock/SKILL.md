---
name: analyze-stock
description: Analyze a specified A-share company from its latest and historical financial reports, apply the valuation models configured by industry in this repository's README.md, and update that company's analysis columns in stocks.md. Use when the user asks to analyze, value, refresh, or review a stock in this value-investing project.
---

# Analyze Stock

Analyze one stock at a time. Treat the work as research support, not personalized investment advice.

## Workflow

1. Resolve the company.
   - Read `stocks.md` and find the row by the user-supplied company name or stock code.
   - Confirm that exactly one row matches. Ask for clarification rather than editing an ambiguous row.
   - Do not add a company or alter the columns through `最新价（元）`; update only the cells to the right of that column unless the user explicitly requests otherwise.

2. Load the repository framework.
   - Read `README.md` before making classifications or calculations.
   - Classify the company using a category defined in `README.md`. State the classification and why it fits.
   - Use that category's specified primary model, cross-validation models, and minimum history requirement.
   - If no applicable category or model exists in `README.md`, do not invent one or copy a model from another category. Report the gap and leave valuation-result cells unchanged. Update only factual fields that can be supported without the missing framework when the user asks for them.

3. Gather source data.
   - Start with the local MCP report archive described in [Local Report MCP](#local-report-mcp). Use its filings as the preferred primary source for A-share periodic reports.
   - Prefer primary disclosures: the issuer's annual, interim, and quarterly reports; exchange announcements; and CNINFO filings for A-share companies.
   - Retrieve the most recently available report and the number of annual reports required by the selected category. Use the newest interim or quarterly report as the forecast anchor when the category requires it.
   - Use a reliable, date-stamped market-data source for the latest share price only if the user asks to refresh `最新价（元）`; otherwise preserve the existing value in `stocks.md`.
   - Record every material source with its report period, publication date, access date, and URL in the response. Distinguish reported figures from estimates.
   - Do not fill gaps with unaudited third-party estimates when a primary filing is available. If data cannot be verified, mark the affected conclusion as unavailable.

4. Normalize the financials.
   - Build a compact historical table covering the required years. Include the inputs needed by the selected models and meaningful per-share figures.
   - Remove or separately disclose non-recurring gains/losses, material asset disposals, impairments, fair-value changes, and other items that would distort sustainable earnings or ROE.
   - State material accounting changes, major share-count changes, and any restatements.
   - Define Bear, Base, and Bull assumptions explicitly. Keep assumptions internally consistent across primary and cross-check models.

5. Value the company.
   - Apply the README primary model first, following its stated inputs and history length.
   - Run every cross-validation model named for the category. State the model, key assumptions, and result in a concise form suitable for `stocks.md`.
   - Compare the results. Explain material divergence and lower confidence when models disagree, source data are incomplete, or assumptions are unusually sensitive.
   - Derive target-price scenarios, safety margin versus the `stocks.md` latest price, and buy/hold/reduce/overvaluation ranges from the modeled outputs. Do not present an unsupported point estimate as certain.
   - Assess Buffett and Graham frameworks separately. Tie each conclusion to durable economics, leverage, earnings quality, capital allocation, valuation, and downside protection rather than a generic label.

6. Update `stocks.md`.
   - Preserve the Markdown table structure, header order, company name, code, total market value, and latest-price cells.
   - Update all applicable cells after `最新价（元）` for the matched row:
     `行业分类`, `企业质量判断`, `正常化年度净利润`, `情景年度净利润预测（Bear/Base/Bull）`, `主情景目标价（Bear/Base/Bull）`, `交叉估值1（模型与结果）`, `交叉估值2（模型与结果）`, `市场隐含预期（增长/ROE）`, `估值可信度（高/中/低）`, `所需安全边际`, `理想买入价`, `安全买入价`, `持有区间`, `减仓价`, `高估区间起点`, `巴菲特框架判断`, `格雷厄姆框架判断`, and `最新财报期末`.
   - Use concise Chinese text. Put units and currency in cells where helpful. Keep scenario order consistently `Bear/Base/Bull`.
   - Do not use Markdown table pipes inside a cell. Use Chinese punctuation or semicolons instead.
   - When a category is unavailable in `README.md`, do not overwrite prior valuation fields with guesses.

7. Validate and report.
   - Confirm the table still has the same number of columns in its header, separator, and edited row.
   - Re-read the edited row and confirm the latest financial-report period agrees with the source.
   - In the final response, summarize the classification, latest report used, main and cross-check valuation ranges, confidence, largest risks, and exactly which row was updated.

## Bank Category

For the `银行` category currently configured in `README.md`:

- Use `PB-ROE（基于可持续 ROE）` as the primary model.
- Cross-check with `Forward PE（基于一致预期）` and `股息折现模型（DDM）`.
- Use ten years of annual history, with the latest three years and most recent quarterly report serving as the forecast anchor.
- Evaluate sustainable ROE using profitability, credit costs, capital adequacy, leverage, asset quality, provisioning, net-interest margin, fee income, and dividend capacity. Do not extrapolate a one-off ROE.
- State PB, sustainable ROE, cost of equity, payout ratio or retention assumptions, and the relevant per-share book value or earnings inputs.

## Local Report MCP

Use these local MCP tools for financial-report discovery and retrieval:

1. Call `query_reports(stock)` with the `stocks.md` code including its market prefix, such as `sh601398`.
   - Use the returned list to identify the latest periodic report and the annual reports needed for the category's historical window.
   - Retain each required `reportId`, report period, and publication date.

2. Call `get_report(reportId)` for every report selected in the prior step.
   - It downloads the CNINFO PDF to the local cache and returns the local file path.
   - Read the returned PDF path with the available PDF-reading tools; do not invent a path or assume a report was downloaded.

3. If `query_reports` does not provide enough historical annual reports for the model's required window, retrieve the missing primary filings from CNINFO or the issuer/exchange and state the shortfall. Do not reduce the required history without disclosing it.

## Quality Bar

- Show calculations sufficiently for another analyst to reproduce the result.
- Use `数据不足` or `未验证` rather than guessing.
- Never claim real-time prices, consensus, or filings without checking a dated source during the current task.
