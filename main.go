package main

import (
	client "github.com/parth-javiya/baymax-bot/whatsappclient"
)

func main() {
	newClient := client.NewClient()

	newClient.Listen(func(msg client.Message) {
		if msg.Text == "Hi" {
			newClient.SendText(msg.From, "Hello from *github*!")
		}
	})
}
