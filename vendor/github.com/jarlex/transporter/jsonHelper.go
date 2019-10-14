package transporter

import (
    "encoding/json"
    "net/http"
)

type JSONError struct {
    Result string `json:"result"`
    Error  string `json:"error"`
}

type JSONErrorMap struct {
    Result string            `json:"result"`
    Error  map[string]string `json:"error"`
}

func RetriveJSONError(w http.ResponseWriter, e error, et string, status int) {
    
    msg := &JSONError{
        Result: et,
        Error:  e.Error(),
    }
    bmsg, _ := json.Marshal(msg)
    WriteJSON(w, status, bmsg)
}

func RetriveJSONErrorMap(w http.ResponseWriter, errorMap map[string]string, et string, status int) {
    
    msg := &JSONErrorMap{
        Result: et,
        Error:  errorMap,
    }
    bmsg, _ := json.Marshal(msg)
    WriteJSON(w, status, bmsg)
}

func WriteJSON(rw http.ResponseWriter, status int, output []byte) {
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(status)
    rw.Write(output)
}
