package tencent

type Store interface {
	SaveKline(kline []Kline, labels map[string]string) error
	SavePlateBoard(board []Plate, labels map[string]string) error
	SaveStockBoard(board []Stock, labels map[string]string) error
}
