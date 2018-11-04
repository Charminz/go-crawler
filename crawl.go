package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/url"
	"github.com/jackdanger/collectlinks"
	"net/http"
	"os"
)

var visited = make(map[string]bool) // map of string -> bool. visited["google.com"] = true and visited["google"]

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Specify URL")
		os.Exit(1)
	}

	// making a queue to keep found links
	queue := make(chan string)
	go func() { queue <- args[0] }()
	for uri := range queue { enqueue(uri, queue) }
}

func connectionSetup() http.Client {
	// &thing{a:b} is like 'new thing(a:b) where thing is object and sets 'a' attribute to a value of 'b'
	tlsConfig := &tls.Config{ InsecureSkipVerify: true }
	transport := &http.Transport{ TLSClientConfig: tlsConfig }
	client := http.Client{ Transport: transport }
	return client
}

func retrieve(uri string, client http.Client, queue chan string) {
	resp, err := client.Get(uri)
	if err != nil { return }

	// We need to close the resource we opened (TCP connection to a web server or a stream of data)
	// defer delays an operation until the function ends
	defer resp.Body.Close()

	// body, _ := ioutil.ReadAll(resp.Body) // resp.Body is a reference to a stream of data. ioutil reads it to memory
	// fmt.Println(string(body)) // casting html to string. by default it is byte array

	links := collectlinks.All(resp.Body)
	for _, link := range links {
		absolute := fixUrl(link, uri)

		if uri != "" {
			if !visited[absolute]{
				go func() { queue <- absolute }()
			}
		}
	}
}

func enqueue(uri string, queue chan string) {
	fmt.Println("Fetching ... ", uri)
	visited[uri] = true
	client := connectionSetup()
	retrieve(uri, client, queue)
}

func fixUrl(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil { return "" }

	baseUrl, err := url.Parse(base)
	if err != nil { return "" }

	uri = baseUrl.ResolveReference(uri)
	return uri.String()
}