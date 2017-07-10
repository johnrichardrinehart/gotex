package main

import (
	"github.com/gorilla/websocket"
	//"net/url"
)

type clients map[*websocket.Conn]*client

type client struct {
	urls []string
	id   int
}
