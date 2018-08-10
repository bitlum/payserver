package crpc

import (
	"runtime"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"fmt"
)

func convertProtoMessage(resp proto.Message) string {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "    ",
		OrigName:     true,
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp)
	if err != nil {
		return fmt.Sprintf("unable to decode response: %v", err)
	}

	return jsonStr
}

func getFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}
