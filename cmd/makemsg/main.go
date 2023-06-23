package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/emersion/go-message"
)

func main() {
	var b bytes.Buffer

	var h message.Header
	h.SetContentType("multipart/alternative", nil)
	h.SetText("To", "lgx@ubikom.cc")
	h.SetText("From", "sot@ubikom.cc")
	h.SetText("Subject", "Updated SMA200/10 SPY strategy")
	w, err := message.CreateWriter(&b, h)
	if err != nil {
		log.Fatal(err)
	}

	var h1 message.Header
	h1.SetContentType("text/html", nil)
	w1, err := w.CreatePart(h1)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		io.WriteString(w1, scanner.Text())
		// fmt.Println(scanner.Text()) // Println will add back the final '\n'
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	w1.Close()

	var h2 message.Header
	h2.SetContentType("text/plain", nil)
	w2, err := w.CreatePart(h2)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w2, "Hello World!\n\nThis is a text part.")
	w2.Close()

	w.Close()

	fmt.Println(b.String())
}
