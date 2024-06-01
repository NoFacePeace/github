package converter

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func Convert(a any, b any) error {
	bs, err := json.Marshal(a)
	if err != nil {
		return errors.New(err.Error())
	}
	if err := json.Unmarshal(bs, b); err != nil {
		return errors.New(err.Error())
	}
	return nil
}
