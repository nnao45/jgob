package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"

	"flag"
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

//Prefix is jgob have route's all information struct
type Prefix struct {
	Remark string `json:"remark"`
	UUID   string `json:"uuid"`
	Age    string `json:"age"`
	Flag   bool   `json:"flag"`
	Attrs  struct {
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

func (p *Prefix) addAsPath(resAry []string) []string {
	if p.Attrs.Aspath != "" {
		s := "aspath " + p.Attrs.Aspath
		resAry = append(resAry, s)
	}
	return resAry
}

func (p *Prefix) addProtocol(resAry []string) []string {
	if p.Attrs.Protocol == "tcp" || p.Attrs.Protocol == "udp" || p.Attrs.Protocol == "icmp" {
		s := "protocol " + p.Attrs.Protocol
		resAry = append(resAry, s)
	}
	return resAry
}

func (p *Prefix) addSrcIP(resAry []string) []string {
	if p.Attrs.Src != "" {
		s := "source " + p.Attrs.Src
		resAry = append(resAry, s)
	}
	return resAry
}

func (p *Prefix) addDstIP(resAry []string) []string {
	if p.Attrs.Dst != "" {
		s := "destination " + p.Attrs.Dst
		resAry = append(resAry, s)
	}
	return resAry
}

func (p *Prefix) addSrcPort(resAry []string) []string {
	if p.Attrs.SrcPort != "" {
		s := "source-port" + p.Attrs.SrcPort
		resAry = append(resAry, s)
	}
	return resAry
}

func (p *Prefix) addDstPort(resAry []string) []string {
	if p.Attrs.DstPort != "" {
		s := "destination-port" + p.Attrs.DstPort
		resAry = append(resAry, s)
	}
	return resAry
}

func (p *Prefix) addOrigin(resAry []string) []string {
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
	return resAry
}

func (p *Prefix) addCommunities(resAry []string) []string {
	if p.Attrs.Communities != "" {
		s := "community " + p.Attrs.Communities
		resAry = append(resAry, s)
	}
	return resAry
}

func (p *Prefix) addExtcomms(resAry []string) []string {
	if p.Attrs.Extcomms == "" {
		s := "then accept"
		resAry = append(resAry, s)
	} else if p.Attrs.Extcomms == "accept" || p.Attrs.Extcomms == "discard" {
		s := "then " + p.Attrs.Extcomms
		resAry = append(resAry, s)
	} else if p.Attrs.Extcomms != "" {
		s := "then rate-limit " + p.Attrs.Extcomms
		resAry = append(resAry, s)
	}
	return resAry
}

func (p *Prefix) addPrefixFunc(resAry []string) []string {
	resAry = p.addAsPath(resAry)
	resAry = p.addProtocol(resAry)
	resAry = p.addSrcIP(resAry)
	resAry = p.addDstIP(resAry)
	resAry = p.addSrcPort(resAry)
	resAry = p.addDstPort(resAry)
	resAry = p.addOrigin(resAry)
	resAry = p.addCommunities(resAry)
	resAry = p.addExtcomms(resAry)
	return resAry
}

func (p *Prefix) addUUID(res string) string {
	if p.UUID != "" {
		res = p.UUID
		return res
	}
	return res
}

func addog(text string, filename string) {
	var writer *bufio.Writer
	data := []byte(text)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
	writer = bufio.NewWriter(f)
	writer.Write(data)
	writer.Flush()
	if err != nil {
		log.Fatal("Error loading " + filename + " file")
	}
	defer f.Close()
}

// EnvLoad is dotenv func
func EnvLoad() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")

	}
}

var configFile = flag.String("f", CONFIG_FILE, "")
var routeFile = flag.String("r", ROUTE_FILE, "")
var usage = `
Usage:
    jgob_IPv4Flowspec [-r route-file] [-f config-file]

Examples:
    jgob_IPv4Flowspec
    jgob_IPv4Flowspec -r test.rib -f tokyo.tml

`

const (
	//CONFIG_FILE is config-file
	CONFIG_FILE = "config.tml"
	//ROUTE_FILE is rib in file format
	ROUTE_FILE = "jgob.route"
	//DEFAULT_TIMEOUT is Between the API communication
	DEFAULT_TIMEOUT = 2000
	//FLAG_FALSE_ROW is false flag row string
	FLAG_FALSE_ROW = `[{"flag":false}]`
	//FLAG_FALSE_ROW_ONE is false flag row string
	FLAG_FALSE_ROW_ONE = `{"flag":false}`
)

func init() {
	flag.Usage = func() {
		fmt.Printf(usage)
	}

	flag.Parse()
}

func main() {

	achan := make(chan []string)
	schan := make(chan string)
	rchan := make(chan string)

	go jgobServer(achan, schan, rchan)

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
		} else {
			str := "JGOB is up and running\n"
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", strconv.Itoa(len(str)))
			w.Write([]byte(str))
		}
	})

	rtr.HandleFunc("/remark", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
		} else {
			RemarkMap["flag"] = true
			json, err := json.Marshal(RemarkMap)
			if err != nil {
				logrus.Error(err)
			}
			str := fmt.Sprintf("[%s]", string(json))
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(str)))
			w.Write([]byte(str))
		}
	})

	rtr.HandleFunc("/route", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
		} else {
			schan <- "route"
			timer := time.NewTimer(DEFAULT_TIMEOUT * time.Millisecond)
			var str string
			for {
				timer.Reset(DEFAULT_TIMEOUT * time.Millisecond)
				select {
				case str = <-rchan:
					break
				case <-timer.C:
					str = FLAG_FALSE_ROW
					break
				}
				break
			}
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
		} else {
			schan <- "global"
			timer := time.NewTimer(DEFAULT_TIMEOUT * time.Millisecond)
			var str string
			for {
				timer.Reset(DEFAULT_TIMEOUT * time.Millisecond)
				select {
				case str = <-rchan:
					break
				case <-timer.C:
					str = FLAG_FALSE_ROW
					break
				}
				break
			}
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
		} else {
			schan <- "nei"
			timer := time.NewTimer(DEFAULT_TIMEOUT * time.Millisecond)
			var str string
			for {
				timer.Reset(DEFAULT_TIMEOUT * time.Millisecond)
				select {
				case str = <-rchan:
					break
				case <-timer.C:
					str = FLAG_FALSE_ROW
					break
				}
				break
			}
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
				logrus.Error(err)
				return
			}

			//Read body data to parse json
			body := make([]byte, length)
			length, err = r.Body.Read(body)
			if err != nil && err != io.EOF {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Error(err)
				return
			}

			//parse json
			var prefixies []Prefix
			err = json.Unmarshal(body[:length], &prefixies)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Error(err)
				return
			}

			reqAry := make([]string, 0, 50)

			for _, p := range prefixies {
				resAry := make([]string, 0, 20)
				resAry = p.addPrefixFunc(resAry)
				achan <- []string{fmt.Sprint("match " + strings.Join(resAry, " ")), p.Remark}
				timer := time.NewTimer(DEFAULT_TIMEOUT * time.Millisecond)
				var str string
				for {
					timer.Reset(DEFAULT_TIMEOUT * time.Millisecond)
					select {
					case str = <-rchan:
						break
					case <-timer.C:
						str = FLAG_FALSE_ROW_ONE
						break
					}
					break
				}
				reqAry = append(reqAry, str)
				time.Sleep(500 * time.Millisecond)
			}
			var reql string
			for i, req := range reqAry {
				if i+1 < len(reqAry) {
					req = req + `,`
				}
				reql = reql + req
			}
			reql = fmt.Sprintf("[%s]", reql)
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
				logrus.Error(err)
				return
			}

			//Read body data to parse json
			body := make([]byte, length)
			length, err = r.Body.Read(body)
			if err != nil && err != io.EOF {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Error(err)
				return
			}

			//parse json
			var prefixies []Prefix
			err = json.Unmarshal(body[:length], &prefixies)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Error(err)
				return
			}

			var res string
			reqAry := make([]string, 0, 50)

			for _, p := range prefixies {
				res = p.addUUID(res)
				achan <- []string{res, ""}
				timer := time.NewTimer(DEFAULT_TIMEOUT * time.Millisecond)
				var str string
				for {
					timer.Reset(DEFAULT_TIMEOUT * time.Millisecond)
					select {
					case str = <-rchan:
						break
					case <-timer.C:
						str = FLAG_FALSE_ROW_ONE
						break
					}
					break
				}
				reqAry = append(reqAry, str)
				time.Sleep(500 * time.Millisecond)
			}
			var reql string
			for i, req := range reqAry {
				if i+1 < len(reqAry) {
					req = req + `,`
				}
				reql = reql + req
			}
			reql = fmt.Sprintf("[%s]", reql)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(reql)))
			w.Write([]byte(reql))
		}
	})

	w := l.Writer()
	defer w.Close()

	http.Handle("/", rtr)

	accessFile, err := os.OpenFile("access_log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer accessFile.Close()

	srv := &http.Server{
		Addr:     ":9443",
		ErrorLog: log.New(w, "", 0),
		Handler:  httpLogger.WriteLog(http.DefaultServeMux, accessFile),
	}
	logrus.Fatal(srv.ListenAndServeTLS("ssl/development/myself.crt", "ssl/development/myself.key"))
}

func checkAuth(r *http.Request) bool {
	EnvLoad()
	username, password, ok := r.BasicAuth()
	if ok == false {
		return false
	}
	return username == os.Getenv("USERNAME") && password == os.Getenv("PASSWORD")
}
