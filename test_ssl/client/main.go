package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	u := "https://localhost/test"
	request, _ := http.NewRequest("GET", u, nil)


	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: "net-gobgp",
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Transport: tr,
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode, string(contents))
}
