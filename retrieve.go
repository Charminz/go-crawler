package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	resp, err := http.Get("http://ttu.ee")
	fmt.Println("http transport error is:", err)

	body, err := ioutil.ReadAll(resp.Body) // resp.Body is a reference to a stream of data. ioutil reads it to memory
	fmt.Println("read error is: ", err)
	fmt.Println(string(body)) // casting html to string. by default it is byte array

}
