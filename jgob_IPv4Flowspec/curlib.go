package main

import (
	"io/ioutil"
	"bytes"
	"net/http"
	"net/url"
	"crypto/tls"
)

func curlCheck(user, pass string) bool {
	//req, err := http.NewRequest("GET", "http://localhost:8080/test", nil)
	req, err := http.NewRequest("GET", "https://localhost/test", nil)
	if err != nil {
		return false
	}
	req.SetBasicAuth(user, pass)

	tr := &http.Transport{
		 TLSClientConfig: &tls.Config{
			ServerName: "net-gobgp",
			InsecureSkipVerify: true,
		},
	}				                 
	client := &http.Client{
		Transport: tr,
	}

	//resp, err := http.DefaultClient.Do(req)
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		return true
	}
	defer resp.Body.Close()
	return false
}

func curlPost(values url.Values, cmd, user, pass string) error {
	jsondata := bytes.NewBuffer([]byte(cmd))
	//req, err := http.NewRequest("POST", "http://localhost:8080/add", jsondata)
	req, err := http.NewRequest("POST", "https://localhost/add", jsondata)
	if err != nil {
		return err
	}
	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: "net-gobgp",
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Transport: tr,
	}

	//resp, err := http.DefaultClient.Do(req)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	execute(resp)

	return nil
}

func execute(resp *http.Response) string {
	b, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		return string(b)
	}
	return ""
}
