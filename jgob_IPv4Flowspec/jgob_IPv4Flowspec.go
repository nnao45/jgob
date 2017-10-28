package main

import (
	"bufio"
	"encoding/json"
	//	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Prefix struct {
	Uuid  string `json:"uuid"`
	Attrs struct {
		Aspath      string `json:"aspath"`
		Protocol    string `json:"protocol"`
		Src         string `json:"source"`
		Dst         string `json:"destination"`
		SrcPort     string `json:"source-port"`
		DstPort     string `json:"destination-port"`
		Origin      string `json:"origin"`
		Communities string `json:"community"`
		Extcomms    string `json:"extcomms"`
	}
}

func addog(text string, filename string) {
	var writer *bufio.Writer
	data := []byte(text)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
	writer = bufio.NewWriter(f)
	writer.Write(data)
	writer.Flush()
	if err != nil {
		log.Fatal("Error loading %s file", filename)
	}
	defer f.Close()
}

func Env_load() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")

	}
}

func main() {

	achan := make(chan string)
	schan := make(chan string)
	rchan := make(chan string)
	//open := make(chan struct{}, 0)

	go JgobServer(achan, schan, rchan)

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {
			str := "JGOB is up and running\n"
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", strconv.Itoa(len(str)))
			w.Write([]byte(str))
		}
	})

	http.HandleFunc("/route", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {
			schan <- "route"
			str := <-rchan
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(str)))
			w.Write([]byte(str))
		}
	})

	http.HandleFunc("/nei", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {
			schan <- "nei"
			str := <-rchan
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(str)))
			w.Write([]byte(str))
		}
	})

	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {
			schan <- "reload"
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Reloding routing table is done.\n"))
		}
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {

			if r.Method != "POST" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			//To allocate slice for request body
			length, err := strconv.Atoi(r.Header.Get("Content-Length"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			//Read body data to parse json
			body := make([]byte, length)
			length, err = r.Body.Read(body)
			if err != nil && err != io.EOF {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			//parse json
			var prefixies []Prefix
			err = json.Unmarshal(body[:length], &prefixies)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res := "match "

			for _, p := range prefixies {
				resAry := make([]string, 0, 20)

				if p.Attrs.Aspath != "" {
					s := "aspath " + p.Attrs.Aspath
					resAry = append(resAry, s)
				}

				if p.Attrs.Protocol == "tcp" || p.Attrs.Protocol == "udp" || p.Attrs.Protocol == "icmp" {
					s := "protocol " + p.Attrs.Protocol
					resAry = append(resAry, s)
				}

				if p.Attrs.Src != "" {
					s := "source " + p.Attrs.Src
					resAry = append(resAry, s)
				}

				if p.Attrs.Dst != "" {
					s := "destination " + p.Attrs.Dst
					resAry = append(resAry, s)
				}

				if p.Attrs.SrcPort != "" {
					s := "source-port" + p.Attrs.SrcPort
					resAry = append(resAry, s)
				}

				if p.Attrs.DstPort != "" {
					s := "destination-port" + p.Attrs.DstPort
					resAry = append(resAry, s)
				}

				if p.Attrs.Origin == " i" {
					p.Attrs.Origin = "igp"
				} else if p.Attrs.Origin == " e" {
					p.Attrs.Origin = "egp"
				} else if p.Attrs.Origin == " ?" {
					p.Attrs.Origin = "incomplete"
				}

				if p.Attrs.Origin == "igp" || p.Attrs.Origin == "egp" || p.Attrs.Origin == "incomplete" {
					s := "origin " + p.Attrs.Origin
					resAry = append(resAry, s)
				}

				if p.Attrs.Communities != "" {
					s := "community " + p.Attrs.Communities
					resAry = append(resAry, s)
				}

				if p.Attrs.Extcomms == "accept" || p.Attrs.Extcomms == "discard" {
					s := "then " + p.Attrs.Extcomms
					resAry = append(resAry, s)
				} else if p.Attrs.Extcomms != "" {
					s := "then rate-limit " + p.Attrs.Extcomms
					resAry = append(resAry, s)
				}

				restr := res + strings.Join(resAry, " ")
				achan <- restr
				time.Sleep(500 * time.Millisecond)
			}
			w.WriteHeader(http.StatusOK)
		}
	})

	http.HandleFunc("/del", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {

			if r.Method != "POST" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			//To allocate slice for request body
			length, err := strconv.Atoi(r.Header.Get("Content-Length"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			//Read body data to parse json
			body := make([]byte, length)
			length, err = r.Body.Read(body)
			if err != nil && err != io.EOF {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			//parse json
			/*var jsonBody map[string]string
			err = json.Unmarshal(body[:length], &jsonBody)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}*/

			var prefixies []Prefix
			err = json.Unmarshal(body[:length], &prefixies)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var res string

			/*
				if jsonBody["uuid"] != "" {
					res = jsonBody["uuid"]
				}*/

			for _, p := range prefixies {
				if p.Uuid != "" {
					res = p.Uuid
				}

				achan <- res
			}
			w.WriteHeader(http.StatusOK)
		}
	})

	//log.Fatal(http.ListenAndServe(":8080", nil))
	log.Fatal(http.ListenAndServeTLS(":443", "ssl/development/myself.crt", "ssl/development/myself.key", nil))
}

func checkAuth(r *http.Request) bool {
	Env_load()
	username, password, ok := r.BasicAuth()
	if ok == false {
		return false
	}
	return username == os.Getenv("USERNAME") && password == os.Getenv("PASSWORD")
}
