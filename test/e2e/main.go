package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	waitUntilWebhookReady()
	fmt.Println("done")
}

func waitUntilWebhookReady() {
	fmt.Print("Waiting for the webhook server")

	for i := 0; i < 100; i++ {
		if _, err := http.Get("http://pullup-webhook"); err == nil {
			fmt.Println("")
			return
		}

		fmt.Print(".")
		time.Sleep(time.Second)
	}
}
