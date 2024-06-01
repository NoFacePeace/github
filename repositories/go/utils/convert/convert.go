package convert

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func StructToMap(v any) (map[string]any, error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	m := map[string]any{}
	if err := json.Unmarshal(bs, &m); err != nil {
		return nil, errors.New(err.Error())
	}
	return m, nil
}
