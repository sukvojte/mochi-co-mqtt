// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 mochi-co
// SPDX-FileContributor: mochi-co

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sukvojte/mochi-co-mqtt"
	"github.com/sukvojte/mochi-co-mqtt/hooks/auth"
	"github.com/sukvojte/mochi-co-mqtt/hooks/storage/bolt"
	"github.com/sukvojte/mochi-co-mqtt/listeners"
	"go.etcd.io/bbolt"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	server := mqtt.New(nil)
	_ = server.AddHook(new(auth.AllowHook), nil)

	err := server.AddHook(new(bolt.Hook), &bolt.Options{
		Path: "bolt.db",
		Options: &bbolt.Options{
			Timeout: 500 * time.Millisecond,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	tcp := listeners.NewTCP("t1", ":1883", nil)
	err = server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-done
	server.Log.Warn().Msg("caught signal, stopping...")
	server.Close()
	server.Log.Info().Msg("main.go finished")
}
