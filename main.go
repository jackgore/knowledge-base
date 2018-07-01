package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/JonathonGore/knowledge-base/config"
	"github.com/JonathonGore/knowledge-base/handlers"
	_ "github.com/JonathonGore/knowledge-base/logging"
	"github.com/JonathonGore/knowledge-base/server"
	"github.com/JonathonGore/knowledge-base/session/managers"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/storage/sql"
)

func main() {
	confFile := flag.String("config", "config.yml", "specify the config file to use")
	flag.Parse()

	var (
		api  handlers.API
		conf config.Config
		d    storage.Driver
	)

	conf, err := config.New(*confFile)
	if err != nil {
		log.Fatalf("unable to parse configuration file: %v", err)
	}

	d, err = sql.New(conf.Database)
	if err != nil {
		log.Fatalf("unable to create sql driver: %v", err)
	}

	sm, err := managers.NewSMManager(conf.CookieName, conf.PublicCookieName, conf.CookieDuration, d)
	if err != nil {
		log.Fatalf("unable to create session manager: %v", err)
	}

	api, err = handlers.New(d, sm)
	if err != nil {
		log.Fatalf("unable to create handler: %v", err)
	}

	s, err := server.New(api, sm, d, conf.AllowPublicQuestions)
	if err != nil {
		log.Fatalf("error initializing server: %v", err)
	}

	srv := &http.Server{
		Addr:      fmt.Sprintf(":%v", conf.Port),
		Handler:   s,
		TLSConfig: &tls.Config{},
	}

	log.Printf("Starting server over http on port: %v", conf.Port)
	log.Fatal(srv.ListenAndServe())
}
