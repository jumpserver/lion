package tunnel

import (
	"encoding/base64"
	"unicode/utf8"

	"lion/pkg/guacd"
	"lion/pkg/session"
)

type clipboardTransfer struct {
	index     string
	text      bool
	allow     bool
	maxLength int
	size      int
}

type clipboardPolicyFilter struct {
	perm     *session.ActionPermission
	toClient map[string]*clipboardTransfer
	toServer map[string]*clipboardTransfer
}

func newClipboardPolicyFilter(perm *session.ActionPermission) *clipboardPolicyFilter {
	return &clipboardPolicyFilter{
		perm:     perm,
		toClient: map[string]*clipboardTransfer{},
		toServer: map[string]*clipboardTransfer{},
	}
}

func (f *clipboardPolicyFilter) filterToClient(instruction *guacd.Instruction) *guacd.Instruction {
	return f.filter(instruction, f.toClient, true)
}

func (f *clipboardPolicyFilter) filterToServer(instruction *guacd.Instruction) *guacd.Instruction {
	return f.filter(instruction, f.toServer, false)
}

// streamPolicy resolves whether a clipboard stream may pass and the max text
// length (0 = unlimited), by direction and content type:
//
//	toClient  (server -> client, copy): text via TextCopy,  file via FileDownload
//	!toClient (client -> server, paste): text via TextPaste, file via FileUpload
func (f *clipboardPolicyFilter) streamPolicy(toClient, text bool) (bool, int) {
	p := f.perm.ClipboardPolicy
	if toClient {
		if text {
			return f.perm.EnableCopy && p.TextCopy, p.TextCopyMaxLength
		}
		return f.perm.EnableCopy && p.FileDownload, 0
	}
	if text {
		return f.perm.EnablePaste && p.TextPaste, p.TextPasteMaxLength
	}
	return f.perm.EnablePaste && p.FileUpload, 0
}

func (f *clipboardPolicyFilter) filter(instruction *guacd.Instruction, transfers map[string]*clipboardTransfer, toClient bool) *guacd.Instruction {
	if f == nil || f.perm == nil || instruction == nil {
		return instruction
	}
	switch instruction.Opcode {
	case guacd.InstructionStreamingClipboard:
		if len(instruction.Args) < 2 {
			return instruction
		}
		text := instruction.Args[1] == "text/plain"
		allow, maxLength := f.streamPolicy(toClient, text)
		transfer := &clipboardTransfer{
			index:     instruction.Args[0],
			text:      text,
			allow:     allow,
			maxLength: maxLength,
		}
		transfers[transfer.index] = transfer
		if !transfer.allow {
			return nil
		}
	case guacd.InstructionStreamingBlob:
		if len(instruction.Args) < 2 {
			return instruction
		}
		transfer := transfers[instruction.Args[0]]
		if transfer == nil {
			return instruction
		}
		if !transfer.allow {
			return nil
		}
		if transfer.text && transfer.maxLength > 0 {
			blob, err := base64.StdEncoding.DecodeString(instruction.Args[1])
			if err != nil {
				return nil
			}
			transfer.size += utf8.RuneCount(blob)
			if transfer.size > transfer.maxLength {
				transfer.allow = false
				return nil
			}
		}
	case guacd.InstructionStreamingEnd:
		if len(instruction.Args) > 0 {
			transfer := transfers[instruction.Args[0]]
			delete(transfers, instruction.Args[0])
			if transfer != nil && !transfer.allow {
				return nil
			}
		}
	}
	return instruction
}
