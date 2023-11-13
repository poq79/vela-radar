package syn

import (
	"errors"
	"github.com/vela-ssoc/vela-radar/port"
)

var ErrorNoSyn = errors.New("no syn support")

var DefaultSynOption = port.Option{
	Rate:    500,
	Timeout: 800,
}
