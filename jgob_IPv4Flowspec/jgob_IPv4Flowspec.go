package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"io"
	//	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func Env_load() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")

	}
}

func main() {

	achan := make(chan string)
	schan := make(chan struct{}, 0)
	rchan := make(chan string)

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

	http.HandleFunc("/show", func(w http.ResponseWriter, r *http.Request) {
                if checkAuth(r) == false {
                        w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
                        w.WriteHeader(401)
                        w.Write([]byte("401 Unauthorized\n"))
                        return
                } else {
			w.Write([]byte("#################################\n"))
                        w.Write([]byte("  show flowspec ipv4 in Gobgpd   \n"))
			w.Write([]byte("#################################\n"))
			schan <- struct{}{}
			w.Write([]byte(<-rchan))
                }
        })

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
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
			var jsonBody map[string]string
			err = json.Unmarshal(body[:length], &jsonBody)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			//var res Announce
			res := "match "

			resAry := make([]string ,0 ,20)

			if jsonBody["proto"] == "tcp" || jsonBody["proto"] == "udp" || jsonBody["proto"] == "icmp"  {
				s := "protocol " + jsonBody["proto"]
				resAry = append(resAry, s)
			}

			if jsonBody["src-ip"] != "" {
				s := "source "  + jsonBody["src-ip"]
				resAry = append(resAry, s)
			}

			if jsonBody["dst-ip"] != "" {
				s := "destination "  + jsonBody["dst-ip"]
				resAry = append(resAry, s)
			}

			if jsonBody["src-port"] != "" {
				s := "source-port " + `==` + jsonBody["src-port"] + ""
				resAry = append(resAry, s)
			}

			if jsonBody["dst-port"] != "" {
                                s := "destination-port " + `==` + jsonBody["dst-port"] + ""
                                resAry = append(resAry, s)
                        }

			if jsonBody["action"] == "accept" ||  jsonBody["action"] == "discard"  {
				s:= "then " + jsonBody["action"]
				resAry = append(resAry, s)
			} else if jsonBody["rate-limit"] != "" {
				s:= "then rate-limit " + jsonBody["rate-limit"]
				resAry = append(resAry, s)
			}

                        if jsonBody["origin"] == "igp" || jsonBody["origin"] == "egp" || jsonBody["origin"] == "incomplete" {
                                s := "origin " + jsonBody["origin"]
                                resAry = append(resAry, s)
                        }

                        if jsonBody["aspath"] != "" {
                                s := "aspath " + jsonBody["aspath"]
                                resAry = append(resAry, s)
                        }

                        if jsonBody["community"] != "" {
                                s := "community " + jsonBody["community"]
                                resAry = append(resAry, s)
                        }

			res = res + strings.Join(resAry, " ")

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
