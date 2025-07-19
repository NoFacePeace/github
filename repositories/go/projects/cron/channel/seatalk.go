package channel

import (
	"fmt"

	"github.com/NoFacePeace/github/repositories/go/external/seatalk"
)

type SeaTalk struct {
	cfg     *SeaTalkConfig
	seatalk seatalk.Interface
}

type SeaTalkConfig struct {
	URL   string
	Group string
}

func NewSeaTalk(c *SeaTalkConfig) *SeaTalk {
	return &SeaTalk{
		cfg:     c,
		seatalk: seatalk.New(c.URL),
	}
}

func (s *SeaTalk) Send(msg *Message) error {
	if err := s.seatalk.SendGroupText(s.cfg.Group, msg.String()); err != nil {
		return fmt.Errorf("seatalk send group text error: [%w]", err)
	}
	return nil
}
