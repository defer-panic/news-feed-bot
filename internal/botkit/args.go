package botkit

import (
	"encoding/json"
)

func ParseJSON[T any](src string) (T, error) {
	var args T

	if err := json.Unmarshal([]byte(src), &args); err != nil {
		return *(new(T)), err
	}

	return args, nil
}
