package finance

import "fmt"

type Plate struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type PlateType string

const (
	PlateTypeHY2 PlateType = "hy2"
)

func ListPlates(typ PlateType, options ...Option) ([]Plate, error) {
	offset := 0
	count := 40
	ps := []Plate{}
	for {
		options = append(options, WithCount(40), WithOffset(offset))
		data, err := getRank(options...)
		if err != nil {
			return nil, fmt.Errorf("finace get rank error: [%w]", err)
		}
		for _, v := range data.RankList {
			p := Plate{}
			p.Code = v.Code
			p.Name = v.Name
			ps = append(ps, p)
		}
		if len(data.RankList) < 40 {
			break
		}
		offset += count
	}
	return ps, nil
}
