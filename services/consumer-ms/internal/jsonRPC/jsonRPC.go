package jsonRPC

const (
	PARSE_ERROR      = -32700
	INVALID_REQUEST  = -32600
	METHOD_NOT_FOUND = -32601
	INVALID_PARAMS   = -32602
	INTERNAL_ERROR   = -32603
)

type JsonRPCRequest struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	Id      int    `json:"id"`
}

type JsonRPCResponse struct {
	JsonRPC string        `json:"jsonrpc"`
	Result  any           `json:"result,omitempty"`
	Error   *JsonRPCError `json:"error,omitempty"`
	Id      int           `json:"id"`
}

type JsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

var currId int = 0

func NewRequest() JsonRPCRequest {
	currId++
	return JsonRPCRequest{JsonRPC: "2.0", Id: currId}
}

func ResponseWithError(id int, number int, err error) JsonRPCResponse {
	return JsonRPCResponse{JsonRPC: "2.0", Id: id, Error: &JsonRPCError{Code: number, Message: err.Error()}}
}

func NewResponse(id int, result any) JsonRPCResponse {
	return JsonRPCResponse{JsonRPC: "2.0", Id: id, Result: result}
}
