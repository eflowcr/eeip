package main

import (
	"fmt"
	"log"

	"github.com/emersion/go-imap/client"
)

func main() {
	log.Println("Connecting to server...")
	c, err := client.DialTLS("mail.eprac.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	defer c.Logout()

	caps, _ := c.Capability()
	log.Println("Capabilities:", caps)

	log.Println("Logging in...")
	if err := c.Login("eitel.rodriguez@eprac.com", "Tuanis1978"); err != nil {
		log.Fatal("Login error:", err)
	}
	log.Println("Logged in successfully!")
}
