package models

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Document struct {
	ID int
	sync.RWMutex
	Content string
	Clients map[*websocket.Conn]bool
}
