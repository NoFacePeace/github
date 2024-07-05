import yfinance as yf

# 获取贵州茅台股票数据
symbol = "603259.SS"
start_date = "2024-06-01"
end_date = "2024-06-29"

data = yf.download(symbol, start=start_date, end=end_date)

apple = yf.Ticker("837046.BS")
print(apple.info)