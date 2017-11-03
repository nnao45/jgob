package main

import (
	"fmt"
	"bufio"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"

	"github.com/ajays20078/go-http-logger"
	"io"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Prefix struct {
	Uuid  string `json:"uuid"`
	Age   string `json:"age"`
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

	go JgobServer(achan, schan, rchan)

	l := logrus.New()
	l.Out = ioutil.Discard
	l.Formatter = &logrus.TextFormatter{
		DisableColors: true,
	}
	hook, err := lSyslog.NewSyslogHook("", "syslog", syslog.LOG_INFO|syslog.LOG_SYSLOG, "jgobd")
	if err == nil {
		l.Hooks.Add(hook)
	} else {
		panic(err)
	}

	rtr := mux.NewRouter()

	rtr.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
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

	rtr.HandleFunc("/route", func(w http.ResponseWriter, r *http.Request) {
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

	rtr.HandleFunc("/global", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {
			schan <- "global"
			str := <-rchan
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(str)))
			w.Write([]byte(str))
		}
	})

	rtr.HandleFunc("/nei", func(w http.ResponseWriter, r *http.Request) {
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

	rtr.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
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

	rtr.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
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

			reqAry := make([]string, 0, 50)

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

				achan <- "match " + strings.Join(resAry, " ")
				reqAry = append(reqAry, <-rchan)
				time.Sleep(500 * time.Millisecond)
			}
			var reql string
			for i, req := range reqAry {
				if i+1 < len(reqAry) {
					req = req + `,`
				}
				reql = reql + req
			}
			reql = fmt.Sprintf("[%s]",  reql)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(reql)))
			w.Write([]byte(reql))
		}
	})

	rtr.HandleFunc("/del", func(w http.ResponseWriter, r *http.Request) {
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

			var res string

			for _, p := range prefixies {
				if p.Uuid != "" {
					res = p.Uuid
				}

				achan <- res
			}
			w.WriteHeader(http.StatusOK)
		}
	})

	w := l.Writer()
	defer w.Close()

	http.Handle("/", rtr)

	access_file_handler, err := os.OpenFile("access_log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer access_file_handler.Close()

	srv := &http.Server{
		Addr:     ":9443",
		ErrorLog: log.New(w, "", 0),
		Handler:  httpLogger.WriteLog(http.DefaultServeMux, access_file_handler),
	}
	logrus.Fatal(srv.ListenAndServeTLS("ssl/development/myself.crt", "ssl/development/myself.key"))
}

func checkAuth(r *http.Request) bool {
	Env_load()
	username, password, ok := r.BasicAuth()
	if ok == false {
		return false
	}
	return username == os.Getenv("USERNAME") && password == os.Getenv("PASSWORD")
}
