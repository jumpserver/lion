package guacd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Instruction struct {
	Opcode       string
	Args         []string
	ProtocolForm string
}

func NewInstruction(opcode string, args ...string) (ret Instruction) {
	ret.Opcode = opcode
	ret.Args = args
	return ret
}

// 构造 `OPCODE,ARG1,ARG2,ARG3,...;` 的格式
func (opt *Instruction) String() string {
	if len(opt.ProtocolForm) > 0 {
		return opt.ProtocolForm
	}
	opt.ProtocolForm = fmt.Sprintf("%d.%s", len(opt.Opcode), opt.Opcode)
	for _, value := range opt.Args {
		opt.ProtocolForm += fmt.Sprintf(",%d.%s", len([]rune(value)), value)
	}
	opt.ProtocolForm += semicolonDelimiter
	return opt.ProtocolForm
}

const (
	semicolonDelimiter = ";"
)

const (
	ByteDotDelimiter       = '.'
	ByteCommaDelimiter     = ','
	ByteSemicolonDelimiter = ';'
)

var (
	ErrInstructionMissSemicolon = errors.New("instruction without semicolon")
	ErrInstructionMissDot       = errors.New("instruction without dot")
	ErrInstructionBadDigit      = errors.New("instruction with bad digit")
	ErrInstructionBadContent    = errors.New("instruction with bad Content")
)

// raw 是以 `;` 为结束符的原生字符串
func ParseInstructionString(raw string) (ret Instruction, err error) {
	if !strings.HasSuffix(raw, semicolonDelimiter) {
		return Instruction{}, fmt.Errorf("%w: %s", ErrInstructionMissSemicolon, raw)
	}
	raw = trimSuffixSemicolonDelimiter(raw)
	rawRune := []rune(raw)
	args := make([]string, 0, 1024)
	var i = 0
	for len(rawRune) > 0 {
		switch rawRune[i] {
		case ByteCommaDelimiter, ByteSemicolonDelimiter:
			// 重置数据
			rawRune = rawRune[i+1:]
			i = 0
			continue
		case ByteDotDelimiter:
			// 解析 LENGTH.VALUE 格式的数据
			dotIndex := rawRune[:i]
			argContentLen, err := strconv.Atoi(string(dotIndex))
			if err != nil {
				return Instruction{}, fmt.Errorf("%w: %s",
					ErrInstructionBadDigit, err.Error())
			}
			if argContentLen > len(rawRune[i+1:]) {
				return Instruction{}, fmt.Errorf("%w: %s",
					ErrInstructionBadContent, raw)
			}
			argContent := string(rawRune[i+1 : argContentLen+i+1])
			args = append(args, argContent)
			rawRune = rawRune[argContentLen+i+1:]
			i = 0
			continue
		default:
			i++
		}
		if i >= len(rawRune) {
			return Instruction{}, fmt.Errorf("%w: %s", ErrInstructionBadContent, raw)
		}
	}
	if len(args) < 1 {
		return Instruction{}, fmt.Errorf("%w: no content", ErrInstructionBadContent)
	}
	return NewInstruction(args[0], args[1:]...), nil
}

func trimSuffixSemicolonDelimiter(content string) string {
	return strings.TrimSuffix(content, semicolonDelimiter)
}
