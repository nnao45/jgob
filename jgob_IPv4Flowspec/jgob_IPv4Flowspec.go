package main

import (
	"bufio"
	"encoding/json"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

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
			w.Write([]byte("JGOB is up and running\n"))
		}
	})

	http.HandleFunc("/route", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {
			w.Write([]byte("▼show flowspec ipv4 in Gobgpd\n"))
			schan <- "route"
			w.Write([]byte(<-rchan))
		}
	})

	http.HandleFunc("/nei", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {
			w.Write([]byte("▼show bgp neighbor flowspec summary\n"))
			schan <- "nei"
			w.Write([]byte(<-rchan))
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

			t := time.Now()
			tf := t.Format("2006-01-02 15:04:05.000")
			logtf := "[" + tf + "](ADDING) '" + string(body[:length]) + "'\n"
			addog(logtf, "jgob.log")

			//parse json
			var jsonBody map[string]string
			err = json.Unmarshal(body[:length], &jsonBody)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res := "match "

			resAry := make([]string, 0, 20)

			if jsonBody["aspath"] != "" {
				s := "aspath " + jsonBody["aspath"]
				resAry = append(resAry, s)
			}

			if jsonBody["protocol"] == "tcp" || jsonBody["protocol"] == "udp" || jsonBody["protocol"] == "icmp" {
				s := "protocol " + jsonBody["protocol"]
				resAry = append(resAry, s)
			}

			if jsonBody["source"] != "" {
				s := "source " + jsonBody["source"]
				resAry = append(resAry, s)
			}

			if jsonBody["destination"] != "" {
				s := "destination " + jsonBody["destination"]
				resAry = append(resAry, s)
			}

			if jsonBody["source-port"] != "" {
				s := "source-port" + strings.Trim(jsonBody["source-port"], "true")
				resAry = append(resAry, s)
			}

			if jsonBody["destination-port"] != "" {
				s := "destination-port" + strings.Trim(jsonBody["destination-port"], "true")
				resAry = append(resAry, s)
			}

			if jsonBody["origin"] == " i" {
				jsonBody["origin"] = "igp"
			} else if jsonBody["origin"] == " e" {
				jsonBody["origin"] = "egp"
			} else if jsonBody["origin"] == " ?" {
				jsonBody["origin"] = "incomplete"
			}

			if jsonBody["origin"] == "igp" || jsonBody["origin"] == "egp" || jsonBody["origin"] == "incomplete" {
				s := "origin " + jsonBody["origin"]
				resAry = append(resAry, s)
			}

			if jsonBody["communities"] != "" {
				s := "community " + jsonBody["communities"]
				resAry = append(resAry, s)
			}

			if jsonBody["extcomms"] == "accept" || jsonBody["extcomms"] == "discard" {
				s := "then " + jsonBody["extcomms"]
				resAry = append(resAry, s)
			} else if jsonBody["extcomms"] != "" {
				s := "then rate-limit " + jsonBody["extcomms"]
				resAry = append(resAry, s)
			}

			res = res + strings.Join(resAry, " ")
			achan <- res

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

			t := time.Now()
			tf := t.Format("2006-01-02 15:04:05.000")
			logtf := "[" + tf + "](DELETE) '" + string(body[:length]) + "'\n"
			addog(logtf, "jgob.log")

			//parse json
			var jsonBody map[string]string
			err = json.Unmarshal(body[:length], &jsonBody)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var res string

			if jsonBody["uuid"] != "" {
				res = jsonBody["uuid"]
			}

			achan <- res

			w.WriteHeader(http.StatusOK)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func checkAuth(r *http.Request) bool {
	Env_load()
	username, password, ok := r.BasicAuth()
	if ok == false {
		return false
	}
	return username == os.Getenv("USERNAME") && password == os.Getenv("PASSWORD")
}
