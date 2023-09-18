package main

import (
	// "net/http"
	managers "eboox/Managers"
	"net/http"
	"os"
	"os/signal"
)

const TEMP_LIBRARY_PATH string = "/home/khairi/Documents/Library"

func main() {

	println("start")
	managers := managers.ManagersInit(TEMP_LIBRARY_PATH)
	// defer managers.Save()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for {
			<-c
			managers.Save()
			os.Exit(0)
		}
	}()

	if err := http.ListenAndServe(":5050", nil); err != nil {
		panic("die")
	}
}
