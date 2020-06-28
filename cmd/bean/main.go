package main

import (
	"log"
	"net/http"
	"os/user"

	"github.com/gorilla/mux"

	"bean/pkg/bear"
	"bean/pkg/core"
	"bean/pkg/meta"
	"bean/pkg/server"

	_ "github.com/mattn/go-sqlite3"
)

const (
	bearDBPathSuffix = "/Library/Group Containers/9K33E3U3T4.net.shinyfrog.bear/Application Data/database.sqlite"
	metaDBPath       = "bean.meta"
)

func main() {
	user, err := user.Current()
	if err != nil {
		log.Fatalf("get current user: %v", err)
	}
	bearDBPath := user.HomeDir + bearDBPathSuffix
	bearStorage, err := bear.Connect(bearDBPath)
	if err != nil {
		log.Fatalf("connect to bear db: %v", err)
	}
	defer bearStorage.Close()

	metaStorage, err := meta.Connect(metaDBPath)
	if err != nil {
		log.Fatalf("connect to meta db: %v", err)
	}
	defer metaStorage.Close()

	core := core.New(metaStorage, bearStorage)
	// TODO: load on ticker
	if err := core.Load(); err != nil {
		log.Fatalf("load data: %v", err)
	}

	srv, err := server.New(core)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", srv.HandleListTopics).Methods("GET")
	router.HandleFunc("/{topic}", srv.HandleNextCard).Methods("GET")
	router.HandleFunc("/reflect", srv.HandleReflect).Methods("POST")

	if err := http.ListenAndServe("localhost:63411", router); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
