package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/emersion/go-message/mail"
)

func main() {
	msg := "From: x@y.com\r\nSubject: Hello\r\nContent-Type: text/plain\r\n\r\nThis is a test."
	r := strings.NewReader(msg)
	mr, err := mail.CreateReader(r)
	if err != nil {
		panic(err)
	}

	var body string
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			break
		}
		
		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := io.ReadAll(p.Body)
            ct, _, _ := h.ContentType()
			body = string(b) + " (" + ct + ")"
		}
	}
	fmt.Printf("Body loop: %q\n", body)
}
