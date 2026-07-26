package guacd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestParseInstructionString(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Instruction
	}{
		{
			name: "opcode and arguments",
			raw:  "1.a,2.bc,3.def,10.helloworld;",
			want: NewInstruction("a", "bc", "def", "helloworld"),
		},
		{
			name: "empty opcode",
			raw:  "0.;",
			want: NewInstruction(""),
		},
		{
			name: "unicode codepoints",
			raw:  "4.echo,3.你好🙂;",
			want: NewInstruction("echo", "你好🙂"),
		},
		{
			name: "delimiters inside values",
			raw:  "4.echo,3.a;b,3.x,y;",
			want: NewInstruction("echo", "a;b", "x,y"),
		},
		{
			name: "empty argument",
			raw:  "4.test,0.;",
			want: NewInstruction("test", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInstructionString(tt.raw)
			if err != nil {
				t.Fatalf("ParseInstructionString() error = %v", err)
			}
			if got.Opcode != tt.want.Opcode || !reflect.DeepEqual(got.Args, tt.want.Args) {
				t.Fatalf("ParseInstructionString() = %#v, want %#v", got, tt.want)
			}
			if got.String() != tt.raw {
				t.Fatalf("Instruction.String() = %q, want %q", got.String(), tt.raw)
			}
		})
	}
}

func TestParseInstructionStringRejectsMalformedInput(t *testing.T) {
	invalidUTF8 := string([]byte{'1', '.', 0xff, ';'})
	tests := []struct {
		name    string
		raw     string
		wantErr error
	}{
		{
			name:    "missing semicolon",
			raw:     "4.test",
			wantErr: ErrInstructionMissSemicolon,
		},
		{
			name:    "missing dot",
			raw:     "4;",
			wantErr: ErrInstructionMissDot,
		},
		{
			name:    "non-numeric length",
			raw:     "x.test;",
			wantErr: ErrInstructionBadDigit,
		},
		{
			name:    "empty length",
			raw:     ".test;",
			wantErr: ErrInstructionBadDigit,
		},
		{
			name:    "bad element terminator",
			raw:     "4.test!,1.x;",
			wantErr: ErrInstructionBadTerminator,
		},
		{
			name:    "truncated content",
			raw:     "9.test;",
			wantErr: io.ErrUnexpectedEOF,
		},
		{
			name:    "multiple instructions",
			raw:     "3.one;3.two;",
			wantErr: ErrInstructionTrailingData,
		},
		{
			name:    "length prefix too long",
			raw:     "000001.a;",
			wantErr: ErrInstructionTooLong,
		},
		{
			name:    "element too long",
			raw:     "8193.;",
			wantErr: ErrInstructionTooLong,
		},
		{
			name:    "invalid UTF-8",
			raw:     invalidUTF8,
			wantErr: ErrInstructionInvalidUTF8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseInstructionString(tt.raw)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseInstructionString() error = %v, want errors.Is(%v)",
					err, tt.wantErr)
			}
		})
	}
}

func TestInstructionDecoderReadsStreamIncrementally(t *testing.T) {
	reader := &oneByteReader{
		data: []byte("4.echo,3.a;b;3.nop;"),
	}
	decoder := NewInstructionDecoder(reader)

	first, err := decoder.ReadInstruction()
	if err != nil {
		t.Fatalf("first ReadInstruction() error = %v", err)
	}
	if first.Opcode != "echo" || !reflect.DeepEqual(first.Args, []string{"a;b"}) {
		t.Fatalf("first ReadInstruction() = %#v", first)
	}

	second, err := decoder.ReadInstruction()
	if err != nil {
		t.Fatalf("second ReadInstruction() error = %v", err)
	}
	if second.Opcode != "nop" || len(second.Args) != 0 {
		t.Fatalf("second ReadInstruction() = %#v", second)
	}

	if _, err = decoder.ReadInstruction(); !errors.Is(err, io.EOF) {
		t.Fatalf("final ReadInstruction() error = %v, want io.EOF", err)
	}
}

func TestInstructionDecoderReportsTruncatedInstruction(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr error
	}{
		{
			name:    "length",
			raw:     "4",
			wantErr: ErrInstructionMissDot,
		},
		{
			name:    "content",
			raw:     "4.echo,3.ab",
			wantErr: ErrInstructionBadContent,
		},
		{
			name:    "terminator",
			raw:     "4.echo",
			wantErr: ErrInstructionBadTerminator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewInstructionDecoder(strings.NewReader(tt.raw))
			_, err := decoder.ReadInstruction()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ReadInstruction() error = %v, want errors.Is(%v)", err, tt.wantErr)
			}
			if !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("ReadInstruction() error = %v, want io.ErrUnexpectedEOF", err)
			}
		})
	}
}

func TestInstructionDecoderPreservesReadError(t *testing.T) {
	readErr := errors.New("reader failed")
	reader := io.MultiReader(strings.NewReader("4."), errorReader{err: readErr})

	_, err := NewInstructionDecoder(reader).ReadInstruction()
	if !errors.Is(err, readErr) {
		t.Fatalf("ReadInstruction() error = %v, want reader error", err)
	}
	if errors.Is(err, ErrInstructionBadContent) {
		t.Fatalf("ReadInstruction() error = %v, reader failure must not be classified as malformed content", err)
	}
}

func TestInstructionDecoderEnforcesLimits(t *testing.T) {
	t.Run("maximum element length", func(t *testing.T) {
		value := strings.Repeat("a", maxInstructionElementCodepoints)
		raw := fmt.Sprintf("%d.%s;", maxInstructionElementCodepoints, value)
		if _, err := NewInstructionDecoder(strings.NewReader(raw)).ReadInstruction(); err != nil {
			t.Fatalf("ReadInstruction() error = %v", err)
		}
	})

	t.Run("maximum elements", func(t *testing.T) {
		raw := strings.Repeat("0.,", maxInstructionElements-1) + "0.;"
		if _, err := NewInstructionDecoder(strings.NewReader(raw)).ReadInstruction(); err != nil {
			t.Fatalf("ReadInstruction() error = %v", err)
		}
	})

	t.Run("too many elements", func(t *testing.T) {
		raw := strings.Repeat("0.,", maxInstructionElements) + "0.;"
		_, err := NewInstructionDecoder(strings.NewReader(raw)).ReadInstruction()
		if !errors.Is(err, ErrInstructionTooManyElements) {
			t.Fatalf("ReadInstruction() error = %v, want ErrInstructionTooManyElements", err)
		}
	})

	t.Run("UTF-8 byte limit", func(t *testing.T) {
		value := strings.Repeat("🙂", maxInstructionElementCodepoints)
		raw := fmt.Sprintf("%d.%s;", maxInstructionElementCodepoints, value)
		_, err := NewInstructionDecoder(strings.NewReader(raw)).ReadInstruction()
		if !errors.Is(err, ErrInstructionTooLong) {
			t.Fatalf("ReadInstruction() error = %v, want ErrInstructionTooLong", err)
		}
	})
}

func TestTunnelReadInstruction(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	tunnel := &Tunnel{
		conn: serverConn,
		rw: bufio.NewReadWriter(
			bufio.NewReader(serverConn),
			bufio.NewWriter(serverConn),
		),
	}

	writeErr := make(chan error, 1)
	go func() {
		defer clientConn.Close()
		for _, fragment := range []string{"4.echo,3.a", ";b;3.", "nop;"} {
			if _, err := clientConn.Write([]byte(fragment)); err != nil {
				writeErr <- err
				return
			}
		}
		writeErr <- nil
	}()

	first, err := tunnel.ReadInstruction()
	if err != nil {
		t.Fatalf("first ReadInstruction() error = %v", err)
	}
	if first.Opcode != "echo" || !reflect.DeepEqual(first.Args, []string{"a;b"}) {
		t.Fatalf("first ReadInstruction() = %#v", first)
	}

	second, err := tunnel.ReadInstruction()
	if err != nil {
		t.Fatalf("second ReadInstruction() error = %v", err)
	}
	if second.Opcode != "nop" || len(second.Args) != 0 {
		t.Fatalf("second ReadInstruction() = %#v", second)
	}

	if err = <-writeErr; err != nil {
		t.Fatalf("writer error = %v", err)
	}
}

func TestTunnelReadInstructionSetsOneDeadlinePerInstruction(t *testing.T) {
	conn := &deadlineCountingConn{
		reader: strings.NewReader("4.echo,5.a;b;c;"),
	}
	tunnel := &Tunnel{
		conn: conn,
		rw: bufio.NewReadWriter(
			bufio.NewReader(conn),
			bufio.NewWriter(conn),
		),
	}

	instruction, err := tunnel.ReadInstruction()
	if err != nil {
		t.Fatalf("ReadInstruction() error = %v", err)
	}
	if instruction.Opcode != "echo" || !reflect.DeepEqual(instruction.Args, []string{"a;b;c"}) {
		t.Fatalf("ReadInstruction() = %#v", instruction)
	}
	if conn.readDeadlineCalls != 1 {
		t.Fatalf("SetReadDeadline() calls = %d, want 1", conn.readDeadlineCalls)
	}
}

func FuzzInstructionRoundTrip(f *testing.F) {
	f.Add("echo", "a;b,c")
	f.Add("", "")
	f.Add("日志", "你好🙂")

	f.Fuzz(func(t *testing.T, opcode, arg string) {
		if !utf8.ValidString(opcode) || !utf8.ValidString(arg) {
			t.Skip()
		}
		if utf8.RuneCountInString(opcode) > maxInstructionElementCodepoints ||
			utf8.RuneCountInString(arg) > maxInstructionElementCodepoints {
			t.Skip()
		}

		instruction := NewInstruction(opcode, arg)
		raw := instruction.String()
		if len(raw) > maxInstructionBytes {
			t.Skip()
		}

		got, err := ParseInstructionString(raw)
		if err != nil {
			t.Fatalf("ParseInstructionString(%q) error = %v", raw, err)
		}
		if got.Opcode != opcode || !reflect.DeepEqual(got.Args, []string{arg}) {
			t.Fatalf("ParseInstructionString(%q) = %#v", raw, got)
		}
	})
}

type oneByteReader struct {
	data []byte
}

func (r *oneByteReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	p[0] = r.data[0]
	r.data = r.data[1:]
	return 1, nil
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

type deadlineCountingConn struct {
	reader            io.Reader
	readDeadlineCalls int
}

func (c *deadlineCountingConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func (c *deadlineCountingConn) Write(p []byte) (int, error) {
	return len(p), nil
}

func (c *deadlineCountingConn) Close() error {
	return nil
}

func (c *deadlineCountingConn) LocalAddr() net.Addr {
	return testAddr("local")
}

func (c *deadlineCountingConn) RemoteAddr() net.Addr {
	return testAddr("remote")
}

func (c *deadlineCountingConn) SetDeadline(time.Time) error {
	return nil
}

func (c *deadlineCountingConn) SetReadDeadline(time.Time) error {
	c.readDeadlineCalls++
	return nil
}

func (c *deadlineCountingConn) SetWriteDeadline(time.Time) error {
	return nil
}

type testAddr string

func (a testAddr) Network() string {
	return string(a)
}

func (a testAddr) String() string {
	return string(a)
}
