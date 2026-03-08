package broker

import (
	"encoding/json"
)

func Publish(topic string, payload any) (err error) {
	var buf []byte
	switch v := payload.(type) {
	case string:
		buf = []byte(v)
	case []byte:
		buf = v
	case nil:
	default:
		buf, err = json.Marshal(v)
		if err != nil {
			return err
		}
	}
	return server.Publish(topic, buf, false, 0)
}
