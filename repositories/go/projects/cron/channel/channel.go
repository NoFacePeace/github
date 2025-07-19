package channel

import (
	"errors"
	"fmt"

	"github.com/NoFacePeace/github/repositories/go/utils/converter"
)

var ErrTypeNotFound = errors.New("type not found")

type Channel interface {
	Send(*Message) error
}

type Message struct {
	Title   string
	Content string
}

func (m *Message) String() string {
	return fmt.Sprintf("title: %s\ncontent\n%s", m.Title, m.Content)
}

func New(typ string, cfg map[string]any) (Channel, error) {

	switch typ {
	case "seatalk":
		c := &SeaTalkConfig{}
		if err := converter.Convert(cfg, c); err != nil {
			return nil, fmt.Errorf("converter convert error: [%w]", err)
		}
		return NewSeaTalk(c), nil
	}
	return nil, ErrTypeNotFound
}
