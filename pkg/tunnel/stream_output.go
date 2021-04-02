package tunnel

import (
	"encoding/base64"
	"fmt"
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
	filter.Lock()
	defer filter.Unlock()
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

//下载文件的对象
type OutStreamResource struct {
	streamIndex string
	mediaType   string // application/octet-stream
	writer      http.ResponseWriter
	done        chan struct{}
}

func (r *OutStreamResource) Wait() {
	<-r.done
}
