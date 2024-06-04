package tencent

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MySQL struct {
	DB *gorm.DB
}

func NewMySQL(db *gorm.DB) *MySQL {
	cli := &MySQL{
		DB: db,
	}
	db.AutoMigrate(&KlineAStockDay{}, &KlinePlateDay{}, &Plate{}, &Stock{})
	return cli
}

func (m *MySQL) SaveKline(kline []Kline, labels map[string]string) error {
	typ := labels["type"]
	switch typ {
	case StockBoardAstock:
		if err := m.SaveAStockKline(kline); err != nil {
			return err
		}
	case PlateBoardHyTwo:
		if err := m.SavePlateKline(kline); err != nil {
			return err
		}
	}
	return nil
}

func (m *MySQL) SavePlateKline(kline []Kline) error {
	arr := []KlinePlateDay{}
	for _, k := range kline {
		arr = append(arr, KlinePlateDay{
			Kline: k,
		})
	}
	for i := 0; i < len(arr); i += 1000 {
		end := i + 1000
		if end > len(arr) {
			end = len(arr)
		}
		if err := m.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(arr[i:end]).Error; err != nil {
			return errors.New(err.Error())
		}
	}
	return nil
}

func (m *MySQL) SaveAStockKline(kline []Kline) error {
	arr := []KlineAStockDay{}
	for _, k := range kline {
		arr = append(arr, KlineAStockDay{
			Kline: k,
		})
	}
	for i := 0; i < len(arr); i += 1000 {
		end := i + 1000
		if end > len(arr) {
			end = len(arr)
		}
		if err := m.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(arr[i:end]).Error; err != nil {
			return errors.New(err.Error())
		}
	}
	return nil
}

func (m *MySQL) SavePlateBoard(arr []Plate, labels map[string]string) error {
	for i := 0; i < len(arr); i += 1000 {
		end := i + 1000
		if end > len(arr) {
			end = len(arr)
		}
		if err := m.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(arr[i:end]).Error; err != nil {
			return errors.New(err.Error())
		}
	}
	return nil
}

func (m *MySQL) SaveStockBoard(arr []Stock, labels map[string]string) error {
	for i := 0; i < len(arr); i += 1000 {
		end := i + 1000
		if end > len(arr) {
			end = len(arr)
		}
		if err := m.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(arr[i:end]).Error; err != nil {
			return errors.New(err.Error())
		}
	}
	return nil
}

type StockBoardAStock struct {
	gorm.Model
	Stock
}

type KlineAStockDay struct {
	gorm.Model
	Kline
}

func (KlineAStockDay) TableName() string {
	return "t_tencent_kline_astock_day"
}

type KlinePlateDay struct {
	gorm.Model
	Kline
}

func (KlinePlateDay) TableName() string {
	return "t_tencent_kline_plate_day"
}

func (Plate) TableName() string {
	return "t_tencent_plates"
}

func (Stock) TableName() string {
	return "t_tencent_stocks"
}
