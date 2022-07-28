package main

import (
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		DB:   2,
	})
	defer rdb.Close()
	server := &server{
		router:    mux.NewRouter(),
		kv:        rdb,
		templates: NewTemplates(),
	}
	server.ConfigureRoutes()
	http.Handle("/", server.router)
	http.ListenAndServe(":8080", nil)

}
