package render

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type (
	ErrorResponse struct {
		Msg    string `json:"msg"`
		Status int    `json:"status"`
	}
)

func ReadJSONOrFail(rw http.ResponseWriter, req *http.Request, out interface{}) error {
	dec := json.NewDecoder(req.Body)
	err := dec.Decode(out)
	if err != nil {
		WriteError(rw, http.StatusBadRequest, err.Error())
	}
	return err
}

func WriteJSON(rw http.ResponseWriter, status int, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		http.Error(rw, `{"error":{"msg": "internal server error", "status": 500}}`, http.StatusInternalServerError)
		return err
	}
	return WriteJSONRaw(rw, status, data)
}

func WriteSuccess(rw http.ResponseWriter, status int, msg string) error {
	return WriteJSON(rw, status, struct {
		OK  bool   `json:"ok"`
		Msg string `json:"msg"`
	}{
		OK:  true,
		Msg: msg,
	})
}

func WriteError(rw http.ResponseWriter, status int, msg string) error {
	return WriteJSON(rw, status, struct {
		Error ErrorResponse `json:"error"`
	}{
		Error: ErrorResponse{Msg: msg, Status: status},
	})
}

func WriteJSONRaw(rw http.ResponseWriter, status int, data []byte) error {
	rw.Header().Add("Content-Type", "application/json; charset=utf-8")
	rw.Header().Add("Content-Length", strconv.Itoa(len(data)))
	rw.WriteHeader(status)
	rw.Write(data)
	return nil
}
