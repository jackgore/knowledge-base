package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/JonathonGore/dots/yaml"
	"github.com/JonathonGore/knowledge-base/config"
	"github.com/JonathonGore/knowledge-base/handlers"
	"github.com/JonathonGore/knowledge-base/server"
)

func sslExists(certPath, keyPath string) bool {
	if certPath == "" || keyPath == "" {
		return false
	}

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return false
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return false
	}

	return true
}

func main() {
	var api handlers.API
	var conf config.Config

	conf, err := yaml.New("config.yml")
	if err != nil {
		log.Fatalf("unable to parse configuration file", err)
	}

	api, err = handlers.New()
	if err != nil {
		log.Fatalf("unable to create handler: %v", err)
	}

	s, err := server.New(api)
	if err != nil {
		log.Fatalf("error initializing server: %v", err)
	}

	srv := &http.Server{
		Addr:      fmt.Sprintf(":%v", conf.GetInt("server.port")),
		Handler:   s,
		TLSConfig: &tls.Config{},
	}

	certFile := conf.GetString("ssl.certificate")
	keyFile := conf.GetString("ssl.key")

	if !sslExists(certFile, keyFile) {
		log.Printf("Starting server over http on port: %v", conf.GetInt("server.port"))
		log.Fatal(srv.ListenAndServe())
	}

	log.Printf("Starting server over https on port: %v", conf.GetInt("server.port"))
	log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
}
