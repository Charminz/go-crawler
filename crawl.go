/**
	Thanks to Jack Danger Canty for providing the tutorial about building a web crawler in Go and giving useful tips
	Blog post - https://jdanger.com/build-a-web-crawler-in-go.html
 */

package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/jackdanger/collectlinks"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var visited = make(map[string]bool) // map of string -> bool. visited["google.com"] = true and visited["google"]
var oldAcronymsFound = make(chan string) // stores urls which have "ttü" in their content


func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Specify URL")
		os.Exit(1)
	}
	// make a new channel, which takes and gives strings
	queue := make(chan string)
	go func() { queue <- args[0] }() // async function
	for uri := range queue { enqueue(uri, queue) }
}

/**
	Enqueue found URI and retrieve its links. Set URI as visited.
 */
func enqueue(uri string, queue chan string) {
	fmt.Println("Fetching ... ", uri)
	visited[uri] = true
	client := connectionSetup()
	retrieve(uri, client, queue)
}

func retrieve(uri string, client http.Client, queue chan string) {
	resp, err := client.Get(uri)
	if err != nil { return }

	// Closing the client connection. Wait until its finished.
	defer resp.Body.Close()

	// read response body as bytes
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	content := string(bodyBytes)

	if strings.Contains(content, "TTÜ") {
		// Open file to append uri
		f, err := os.OpenFile("oldies.txt", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {panic(err)}

		defer f.Close()

		// Add current URI to file
		if _, err = f.WriteString(uri + "\n"); err != nil {panic(err)}
	}

	// set the response body to original, unread state
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	links := collectlinks.All(resp.Body)

	for _, link := range links {
		absolute := fixUrl(link, uri)
		if uri != "" && !visited[absolute] && strings.Contains(absolute, "ttu.ee"){
			go func() { queue <- absolute }()
		}
	}
}

func connectionSetup() http.Client {
	// &thing{a:b} is like 'new thing(a:b) where thing is object and sets 'a' attribute to a value of 'b'
	tlsConfig := &tls.Config{ InsecureSkipVerify: true }
	transport := &http.Transport{ TLSClientConfig: tlsConfig }
	client := http.Client{ Transport: transport }
	return client
}

func fixUrl(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil { return "" }

	baseUrl, err := url.Parse(base)
	if err != nil { return "" }

	uri = baseUrl.ResolveReference(uri)
	return uri.String()
}