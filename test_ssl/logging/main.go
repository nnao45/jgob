package main

import (
	"log"
	"net/http"

	"fmt"
	accesslog "github.com/mash/go-accesslog"
)

type logger struct {
}

func (l logger) Log(record accesslog.LogRecord) {
	log.Println("IP:" + record.Ip + " User:" + record.Username + " Status:" + fmt.Sprint(record.Status) + " Method:" + record.Method + " Uri:" + record.Uri)
}

type AppHandler struct {
	appName string
}

func (index *AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, %s!", index.appName)
}

func main() {
	l := logger{}
	index := new(AppHandler)
	index.appName = "sample app"
	handler := index
	http.ListenAndServe(":8080", accesslog.NewLoggingHandler(handler, l))
}
