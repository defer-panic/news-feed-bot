package botkit

import "encoding/json"

type CallbackRequest struct {
	Procedure string          `json:"p"`
	Data      json.RawMessage `json:"d"`
}

func MarshalCallbackRequest(procedure string, data any) ([]byte, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	v := CallbackRequest{
		Procedure: procedure,
		Data:      dataJSON,
	}

	return json.Marshal(v)
}

func UnmarshalCallbackRequest[T any](data []byte) (T, error) {
	var v CallbackRequest
	if err := json.Unmarshal(data, &v); err != nil {
		return *(new(T)), err
	}

	var unmarshalledData T
	if err := json.Unmarshal(v.Data, &unmarshalledData); err != nil {
		return *(new(T)), err
	}

	return unmarshalledData, nil
}
