import yfinance as yf
import pandas as pd
import numpy as np
from sklearn.preprocessing import MinMaxScaler
from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import LSTM, Dense, Dropout
import matplotlib.pyplot as plt
import matplotlib.dates as mdates

# 1. 数据收集
ticker = '603259.SS'
data = yf.download(ticker, start='2010-01-01')
data = data[['Close', 'Volume']]

# 2. 数据预处理
scaler = MinMaxScaler(feature_range=(0, 1))
scaled_data = scaler.fit_transform(data)

train_data_length = int(len(scaled_data) * 0.8)
train_data = scaled_data[:train_data_length]
test_data = scaled_data[train_data_length:]

def create_dataset(dataset, time_step=1):
    X, Y = [], []
    for i in range(len(dataset) - time_step - 1):
        X.append(dataset[i:(i + time_step)])
        Y.append(dataset[i + time_step, 0])  # 只预测收盘价
    return np.array(X), np.array(Y)

time_step = 100
X_train, Y_train = create_dataset(train_data, time_step)
X_test, Y_test = create_dataset(test_data, time_step)

X_train = X_train.reshape(X_train.shape[0], X_train.shape[1], X_train.shape[2])
X_test = X_test.reshape(X_test.shape[0], X_test.shape[1], X_test.shape[2])

# 3. 构建 LSTM 模型
model = Sequential()
model.add(LSTM(units=50, return_sequences=True, input_shape=(time_step, X_train.shape[2])))
model.add(LSTM(units=50, return_sequences=False))
model.add(Dense(units=25))
model.add(Dense(units=1))

model.compile(optimizer='adam', loss='mean_squared_error')

# 4. 训练模型
model.fit(X_train, Y_train, batch_size=1, epochs=1)

# 5. 进行预测
train_predict = model.predict(X_train)
test_predict = model.predict(X_test)

train_predict = scaler.inverse_transform(np.concatenate((train_predict, np.zeros((train_predict.shape[0], 1))), axis=1))[:, 0]
test_predict = scaler.inverse_transform(np.concatenate((test_predict, np.zeros((test_predict.shape[0], 1))), axis=1))[:, 0]

# 预测未来 7 天的收盘价
last_100_days = test_data[-time_step:]
future_predictions = []
for _ in range(7):
    last_100_days = last_100_days.reshape(1, time_step, last_100_days.shape[1])
    next_day_prediction = model.predict(last_100_days)
    future_predictions.append(next_day_prediction[0, 0])
    next_day_data = np.array([next_day_prediction[0, 0], last_100_days[0, -1, 1]])  # 添加预测的收盘价和前一天的交易量
    last_100_days = np.append(last_100_days[:, 1:, :], next_day_data.reshape(1, 1, 2), axis=1)

future_predictions = scaler.inverse_transform(np.array(future_predictions).reshape(-1, 1))

# 6. 可视化结果
train_data_plot = np.empty_like(scaled_data)
train_data_plot[:, :] = np.nan
train_data_plot[time_step:len(train_predict)+time_step, 0] = train_predict

test_data_plot = np.empty_like(scaled_data)
test_data_plot[:, :] = np.nan
test_data_plot[len(train_predict)+(time_step*2)+1:len(scaled_data)-1, 0] = test_predict

# 将未来 7 天的预测值添加到测试集预测中
future_dates = pd.date_range(data.index[-1], periods=8, closed='right')
future_data_plot = np.empty((len(scaled_data) + 7, 1))
future_data_plot[:] = np.nan
future_data_plot[:len(scaled_data)] = scaled_data
future_data_plot[len(scaled_data):] = future_predictions

# 绘制实际价格和预测价格
plt.figure(figsize=(12, 6))
plt.plot(data.index, scaler.inverse_transform(scaled_data)[:, 0], label='Actual Price')
plt.plot(data.index[:len(train_data_plot)], train_data_plot[:, 0], label='Training Prediction')
plt.plot(data.index[len(train_data_plot) + time_step:], test_data_plot[:, 0], label='Testing Prediction')
plt.plot(future_dates, future_predictions, label='Future 7 Day Prediction', linestyle='--')

# 设置日期格式
plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m'))
plt.gca().xaxis.set_major_locator(mdates.MonthLocator(interval=6))
plt.gcf().autofmt_xdate()  # 自动旋转日期标签

plt.legend()
plt.title('Stock Price Prediction for 603259.SS')
plt.xlabel('Date')
plt.ylabel('Stock Price')
plt.show()
