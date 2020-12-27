package main

import (
	client "github.com/parth-javiya/baymax-bot/whatsappclient"
)

func main() {
	client := client.NewClient()

	client.Listen(func(msg client.Message) {
		if msg.Text == "Hi" {
			client.SendText(msg.From, "Hello from *github*!")
		}
	})
}
