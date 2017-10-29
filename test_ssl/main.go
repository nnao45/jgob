package main

import (
	"github.com/gorilla/handlers"
	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"log"
	"log/syslog"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		str := "JGOB is up and running\n"
		w.Write([]byte(str))
	})

	l := logrus.New()
	l.Formatter = &logrus.TextFormatter{
		DisableColors: true,
	}
	hook, _ := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO|syslog.LOG_SYSLOG, "test")
	l.Hooks.Add(hook)

	w := l.Writer()
	defer w.Close()
	srv := &http.Server{
		Addr:     ":9443",
		ErrorLog: log.New(w, "", 0),
	}
	logrus.Fatal(srv.ListenAndServeTLS("ssl/development/myself.crt", "ssl/development/myself.key"))
}
