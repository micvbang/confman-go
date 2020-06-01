package httphelpers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// WriteJSON JSON marshals v and writes the result to w.
func WriteJSON(w http.ResponseWriter, v interface{}) error {
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Add("Content-Type", "application/json")

	if _, err := w.Write(bs); err != nil {
		return err
	}

	return nil
}

// WriteJSONRaw writes raw json ([]byte) to w.
func WriteJSONRaw(w http.ResponseWriter, bs []byte) error {
	w.Header().Add("Content-Type", "application/json")

	if _, err := w.Write(bs); err != nil {
		return err
	}

	return nil
}

// ParseJSON reads the body of r and unmarshals it from JSON into v.
func ParseJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf, v)
	if err != nil {
		return err
	}

	return nil
}
