package handlers

import (
	"os"
	"server/internal/rdb"
)

type Handler struct {
	RedisDB rdb.DB
}

var rdbAddr = os.Getenv("rdbaddr")

func NewHand() *Handler {
	rdb, err := rdb.NewClient(rdbAddr)
	if err != nil {
		panic(err)
	}

	return &Handler{
		RedisDB: rdb,
	}
}

type TLSHandler struct{}

func NewTLSHand() *TLSHandler {
	return &TLSHandler{}
}
