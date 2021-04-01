package tunnel

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"guacamole-client-go/pkg/guacd"
)

type OutputStreamInterceptingFilter struct {
	tunnel  *TunnelConn
	streams map[string]OutStreamResource
	sync.Mutex
	acknowledgeBlobs bool
}

func (filter *OutputStreamInterceptingFilter) Filter(unfilteredInstruction *guacd.Instruction) *guacd.Instruction {

	switch unfilteredInstruction.Opcode {
	case guacd.InstructionStreamingAck:
		fmt.Println(unfilteredInstruction.String())
		return unfilteredInstruction
	case guacd.InstructionStreamingArgv:
		fmt.Println(unfilteredInstruction.String())
		return unfilteredInstruction
	case guacd.InstructionStreamingFile:
		fmt.Println(unfilteredInstruction.String())
		return unfilteredInstruction
	case guacd.InstructionStreamingPipe:
		fmt.Println(unfilteredInstruction.String())
		return unfilteredInstruction
	case guacd.InstructionStreamingBlob:
		// Intercept "blob" instructions for in-progress streams
		//if (instruction.getOpcode().equals("blob"))
		//return handleBlob(instruction);
		return filter.handleBlob(unfilteredInstruction)
	case guacd.InstructionStreamingEnd:
		//// Intercept "end" instructions for in-progress streams
		//if (instruction.getOpcode().equals("end")) {
		//	handleEnd(instruction);
		//	return instruction;
		//}
		filter.handleEnd(unfilteredInstruction)
		return unfilteredInstruction
	case guacd.InstructionClientSync:
		// Monitor "sync" instructions to ensure the client does not starve
		// from lack of graphical updates
		//if (instruction.getOpcode().equals("sync")) {
		//	handleSync(instruction);
		//	return instruction;
		//}
		filter.handleSync(unfilteredInstruction)
		return unfilteredInstruction
	case guacd.InstructionObjectPut, guacd.InstructionObjectBody,
		guacd.InstructionObjectGet, guacd.InstructionObjectFilesystem,
		guacd.InstructionObjectUndefine:
		fmt.Println(unfilteredInstruction.String())
		return unfilteredInstruction
	}

	//// Pass instruction through untouched
	//return instruction
	return unfilteredInstruction
}

func (filter *OutputStreamInterceptingFilter) handleBlob(unfilteredInstruction *guacd.Instruction) *guacd.Instruction {
	// Verify all required arguments are present
	args := unfilteredInstruction.Args
	if len(args) < 2 {
		fmt.Println("less two args ", args)
		return unfilteredInstruction
	}
	index := args[0]
	if stream, ok := filter.streams[index]; ok {
		// Decode blob
		data := args[1]
		blob, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if _, err = stream.writer.Write(blob); err != nil {
			filter.sendAck(index, "FAIL", guacd.StatusServerError)
		} else {
			filter.sendAck(index, "OK", guacd.StatusSuccess)
		}

		if !filter.acknowledgeBlobs {
			filter.acknowledgeBlobs = true
			ins := guacd.NewInstruction(guacd.InstructionStreamingBlob, index, "")
			return &ins
		}

		err = filter.sendAck(index, "Ok", guacd.StatusSuccess)
		if err != nil {
			fmt.Println(err)
		}
		return nil
	}
	return unfilteredInstruction
}

func (filter *OutputStreamInterceptingFilter) handleSync(unfilteredInstruction *guacd.Instruction) {
	filter.acknowledgeBlobs = false
}

func (filter *OutputStreamInterceptingFilter) handleEnd(unfilteredInstruction *guacd.Instruction) {
	// Verify all required arguments are present
	//List<String> args = instruction.getArgs();
	//if (args.size() < 1)
	//return;

	// Terminate stream
	//closeInterceptedStream(args.get(0));
	args := unfilteredInstruction.Args
	if len(args) < 1 {
		return
	}
	filter.closeInterceptedStream(args[0])
}

func (filter *OutputStreamInterceptingFilter) sendAck(index, msg string, status guacd.GuacamoleStatus) error {

	// Error "ack" instructions implicitly close the stream
	//if (status != GuacamoleStatus.SUCCESS)
	//	closeInterceptedStream(index);
	//
	//sendInstruction(new GuacamoleInstruction("ack", index, message,
	//	Integer.toString(status.getGuacamoleStatusCode())));
	if status.HttpCode != guacd.StatusSuccess.HttpCode {
		filter.closeInterceptedStream(index)
	}
	return filter.tunnel.WriteTunnelMessage(guacd.NewInstruction(
		guacd.InstructionStreamingAck, index, msg,
		strconv.Itoa(status.GuaCode)))

}

func (filter *OutputStreamInterceptingFilter) closeInterceptedStream(index string) {
	if outStream, ok := filter.streams[index]; ok {
		fmt.Println("closeInterceptedStream index ", index)
		close(outStream.done)
	}
	delete(filter.streams, index)
}

func (filter *OutputStreamInterceptingFilter) addOutStream(out OutStreamResource) {
	filter.Lock()
	defer filter.Unlock()
	filter.streams[out.streamIndex] = out
	err := filter.sendAck(out.streamIndex, "OK", guacd.StatusSuccess)
	if err != nil {
		fmt.Println(err)
	}
}

type InputStreamInterceptingFilter struct {
	tunnel  *TunnelConn
	streams map[string]*InputStreamResource
	sync.Mutex
	acknowledgeBlobs bool
}

func (filter *InputStreamInterceptingFilter) Filter(unfilteredInstruction *guacd.Instruction) *guacd.Instruction {
	filter.Lock()
	defer filter.Unlock()
	if unfilteredInstruction.Opcode == guacd.InstructionStreamingAck {
		filter.handleAck(unfilteredInstruction)
	}
	return unfilteredInstruction
}

func (filter *InputStreamInterceptingFilter) handleAck(unfilteredInstruction *guacd.Instruction) {
	//io.Copy()

	// Verify all required arguments are present
	args := unfilteredInstruction.Args
	if len(args) < 3 {
		return
	}
	index := args[0]
	if stream, ok := filter.streams[index]; ok {
		status := args[2]
		if status != "0" {

			return
		}

		// Send next blob
		filter.readNextBlob(stream)

		//stream.reader.Read()
	}
	return
}

func (filter *InputStreamInterceptingFilter) readNextBlob(stream *InputStreamResource) {
	buf := make([]byte, 6048)
	nr, err := stream.reader.Read(buf)
	if nr > 0 {
		filter.sendBlob(stream.streamIndex, buf[:nr])
	}
	if err != nil {
		if err != io.EOF {
			stream.err = err
		}
		filter.closeInterceptedStream(stream.streamIndex)
		return
	}

}

func (filter *InputStreamInterceptingFilter) sendBlob(index string, p []byte) {
	err := filter.tunnel.WriteTunnelMessage(guacd.NewInstruction(
		guacd.InstructionStreamingBlob, index, base64.StdEncoding.EncodeToString(p)))
	if err != nil {
		fmt.Println(err)
	}
}

func (filter *InputStreamInterceptingFilter) closeInterceptedStream(index string) {
	if outStream, ok := filter.streams[index]; ok {
		fmt.Println("closeInterceptedStream index ", index)
		close(outStream.done)
	}
	delete(filter.streams, index)
}

func (filter *InputStreamInterceptingFilter) sendEnd(index string) {
	err := filter.tunnel.WriteTunnelMessage(guacd.NewInstruction(
		guacd.InstructionStreamingEnd, index))
	if err != nil {
		fmt.Println(err)
	}
}

func (filter *InputStreamInterceptingFilter) addInputStream(stream *InputStreamResource) {
	filter.Lock()
	defer filter.Unlock()
	filter.streams[stream.streamIndex] = stream
	filter.readNextBlob(stream)
}

type OutStreamResource struct {
	streamIndex string
	mediaType   string // application/octet-stream
	writer      http.ResponseWriter
	done        chan struct{}
}

func (o *OutStreamResource) Wait() {
	<-o.done
}

type InputStreamResource struct {
	streamIndex string
	mediaType   string // application/octet-stream
	reader      io.ReadCloser
	done        chan struct{}

	err error
}

func (o *InputStreamResource) Wait() {
	<-o.done
}

func (o *InputStreamResource) WaitErr() error {
	return o.err
}
