package internal

import "encoding/json"

func map2struct[T any](v map[string]any) (*T, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var t T
	err = json.Unmarshal(buf, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func struct2map(s any) (map[string]any, error) {
	buf, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var t map[string]any
	err = json.Unmarshal(buf, &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}
