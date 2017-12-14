package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

/*https://golang.org/pkg/net/http/#Client*/
var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName:         "net-gobgp",
			InsecureSkipVerify: true,
		},
	},
}

func curlCheck(user, pass string) bool {
	//req, err := http.NewRequest("GET", "http://localhost:8080/test", nil)
	req, err := http.NewRequest("GET", "https://localhost:9443/test", nil)
	if err != nil {
		return false
	}
	req.SetBasicAuth(user, pass)

	/*tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName:         "net-gobgp",
			InsecureSkipVerify: true,
		},
	}
	/*client := &http.Client{
		Transport: tr,
	}*/

	//resp, err := http.DefaultClient.Do(req)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
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
	req, err := http.NewRequest("POST", "https://localhost:9443/add", jsondata)
	if err != nil {
		return err
	}
	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")

	/*tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName:         "net-gobgp",
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Transport: tr,
	}*/

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_ = execute(resp)

	return nil
}

func execute(resp *http.Response) string {
	b, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		return string(b)
	}
	return ""
}
