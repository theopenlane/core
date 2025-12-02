package convert

import "encoding/json"

func JSON[T any](in any) (T, error) {
	var out T
	b, err := json.Marshal(in)

	if err != nil {
		return out, err
	}

	if err := json.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}
