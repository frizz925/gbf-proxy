package http

import (
	"log"
	"net/http"
)

func WriteServerError(w http.ResponseWriter, code int, message string, err error) {
	log.Println(err)
	WriteError(w, code, message)
}

func WriteError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, err := w.Write([]byte(message + "\r\n"))
	if err != nil {
		panic(err)
	}
}
