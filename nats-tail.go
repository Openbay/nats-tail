// Copyright(c) 2016 Waldemar Quevedo (waldemar.quevedo@gmail.com)

package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	defaultPaddingSize          = 20
	defaultTimestampPaddingSize = 30
	version                     = "0.1.2"
)

type Engine struct {
	format         string
	longestSubSize int
	showTimestamp  bool
}

func (e *Engine) display(m *nats.Msg) {
	subjectSize := len(m.Subject)
	if subjectSize > e.longestSubSize {
		e.longestSubSize = subjectSize
	}

	log.Println(fmt.Sprintf("%15s %-65s %.500s", time.Now().Format(time.RFC3339), hashColor(m.Subject), string(m.Data)))
}

func NewDefaultEngine(outputFormat string, showTimestamp bool) *Engine {
	return &Engine{
		longestSubSize: defaultPaddingSize,
		format:         outputFormat,
		showTimestamp:  showTimestamp,
	}
}

// hash takes a string and returns it colorized.
func hashColor(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	i := h.Sum32() % 6
	return fmt.Sprintf("\033[1;3%dm%s\033[0m", i+1, s)
}

// Use tls scheme for TLS, e.g. nats-tail -s tls://demo.nats.io:4443 "docker.>"
func usage() {
	log.Fatalf("Usage: nats-tail [-s server] <subject> \n")
}

func main() {
	var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
	var outputFormat = flag.String("o", "docker-logs", "Display output format")
	var showTimestamp = flag.Bool("t", false, "Display timestamp")
	var showVersion = flag.Bool("v", false, "Show nats-tail version")

	token := os.Getenv("NATS_TOKEN")
	if token == "" {
		log.Fatalf("ENV NATS_TOKEN Not found")
	}

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		log.Printf("nats-tail v%s", version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	nc, err := nats.Connect(*urls, nats.Token(token))
	if err != nil {
		log.Fatalf("Can't connect: %s\n", err)
	}

	engine := NewDefaultEngine(*outputFormat, *showTimestamp)
	subj := args[0]
	nc.Subscribe(subj, func(msg *nats.Msg) {
		engine.display(msg)
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on [%s]\n", subj)

	runtime.Goexit()
}
