package testutil

import "encoding/json"

func MustMarshalJSON(v interface{}) []byte {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return buf
}
