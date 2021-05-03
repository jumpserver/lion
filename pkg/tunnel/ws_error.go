package tunnel

import (
	"strconv"

	"lion/pkg/guacd"
)

type GuacamoleServerError struct {
	err  error
	code int
}

func (g GuacamoleServerError) String() guacd.Instruction {
	return guacd.NewInstruction(
		guacd.InstructionServerError,
		g.err.Error(),
		strconv.Itoa(g.code))
}
