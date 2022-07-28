package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type server struct {
	kv        *redis.Client
	router    *mux.Router
	templates *Templates
}
