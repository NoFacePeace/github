import yfinance as yf
import numpy as np

from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import Dense, Dropout, LSTM, Bidirectional
from sklearn.preprocessing import MinMaxScaler


symbol = "603259.SS"
stock = yf.Ticker(symbol)
data = stock.history(period="max")
close = data["Close"].values.reshape(-1, 1)
scaler = MinMaxScaler(feature_range=(0, 1))
close = scaler.fit_transform(close)
print(close.shape)
training_size = round(len(close) * 0.80)
train_data = close[:training_size]
test_data  = close[training_size:]
print(train_data.shape, test_data.shape)

def create_dataset(dataset, time_step=1):
    dataX, dataY = [], []
    for i in range(len(dataset) - time_step - 1):
        a = dataset[i:(i + time_step), 0]
        dataX.append(a)
        dataY.append(dataset[i + time_step, 0])
    return np.array(dataX), np.array(dataY)
train_seq, train_label = create_dataset(train_data)
test_seq, test_label = create_dataset(test_data)
print(train_seq.shape, train_label.shape, test_seq.shape, test_label.shape)

model = Sequential()
model.add(LSTM(units=50, return_sequences=True, input_shape = (train_seq.shape[1], train_seq.shape[2])))

model.add(Dropout(0.1)) 
model.add(LSTM(units=50))

model.add(Dense(2))

model.compile(loss='mean_squared_error', optimizer='adam', metrics=['mean_absolute_error'])

model.summary()