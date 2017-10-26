package main

import "net/http"

func main() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		str := "JGOB is up and running\n"
		w.Write([]byte(str))
	})

	_ = http.ListenAndServeTLS(":443", "ssl/development/myself.crt", "ssl/development/myself.key", nil)
}
