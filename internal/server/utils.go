package server

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func writeError(writer http.ResponseWriter, err error) {
	_, err = writer.Write([]byte(err.Error()))
	if err != nil {
		log.Println("write error err:", err)
	}
}
