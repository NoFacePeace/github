package tool

import "encoding/json"

type Call struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}
