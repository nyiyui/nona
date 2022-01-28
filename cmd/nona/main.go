package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"nyiyui.ca/nona/pkg/nona"
)

type Config struct {
	Store struct {
		Type   string          `json:"type"`
		Config json.RawMessage `json:"config"`
	} `json:"store"`
	Log struct {
		Type   string          `json:"type"`
		Config json.RawMessage `json:"config"`
	} `json:"log"`
	Server nona.Config `json:"server"`
}

func main2() (err error) {
	var configPath string
	var addr string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&addr, "addr", ":8080", "address to listen on")
	flag.Parse()

	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := file.Close(); err2 != nil {
			err = err2
		}
	}()
	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return err
	}
	store1, ok := nona.Stores[config.Store.Type]
	if !ok {
		return fmt.Errorf("store type %s not found", config.Store.Type)
	}
	store2, err := store1(config.Store.Config)
	if err != nil {
		return err
	}
	log.Printf("using store %T", store2)
	log1, ok := nona.Logs[config.Log.Type]
	if !ok {
		return fmt.Errorf("log type %s not found", config.Log.Type)
	}
	log2, err := log1(config.Log.Config)
	if err != nil {
		return err
	}
	log.Printf("using log %T", log2)

	s2, err := nona.NewServer(store2, log2, config.Server)
	if err != nil {
		return err
	}
	return http.ListenAndServe(addr, s2)
}

func main() {
	if err := main2(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
