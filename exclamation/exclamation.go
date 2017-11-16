package main

import (
	"github.com/ajays20078/go-http-logger"
	"github.com/codeskyblue/go-sh"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"strconv"
)

func main() {

	EnvLoad()

	l := logrus.New()
	l.Out = ioutil.Discard
	l.Formatter = &logrus.TextFormatter{
		DisableColors: true,
	}
	hook, err := lSyslog.NewSyslogHook("", "syslog", syslog.LOG_INFO|syslog.LOG_SYSLOG, "exclamation")
	if err == nil {
		l.Hooks.Add(hook)
	} else {
		panic(err)
	}

	rtr := mux.NewRouter()

	rtr.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="EXCLA REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
		} else {
			str := "This is exclamation point!!\n"
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", strconv.Itoa(len(str)))
			w.Write([]byte(str))
		}
	})

	rtr.HandleFunc("/exclamation", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="EXCLA REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
		} else {
			sh.Command("iptables", "-I", "INPUT", "1", "-p", "tcp", "-j", "REJECT", "-m", "tcp", "--dport", "179").Run()
			str := "Reject BGP Commection\n"
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", strconv.Itoa(len(str)))
			w.Write([]byte(str))
		}
	})

	w := l.Writer()
	defer w.Close()
	accessFile, err := os.OpenFile("access_log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer accessFile.Close()

	srv := &http.Server{
		Addr:     ":39443",
		ErrorLog: log.New(w, "", 0),
		Handler:  httpLogger.WriteLog(http.DefaultServeMux, accessFile),
	}

	l.Info("exclamation point started...")

	l.Fatal(srv.ListenAndServeTLS("ssl/development/myself.crt", "ssl/development/myself.key"))
}

func EnvLoad() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")

	}
}

func checkAuth(r *http.Request) bool {
	EnvLoad()
	username, password, ok := r.BasicAuth()
	if ok == false {
		return false
	}
	return username == os.Getenv("USERNAME") && password == os.Getenv("PASSWORD")
}
