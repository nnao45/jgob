package main

import (
	"github.com/joho/godotenv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Announce struct {
	ORIGIN  uint8    `json:"origin"`  //bgp origin-as attr.
	NEXTHOP string   `json:"nexthop"` //bgp nexthop attr.
	ASPATH  []uint32 `json:"aspath"`  //bgp as-path attr.
	SRC_IP  string   `json:"src-ip"`  //bgp announce prefix of address.
	SRC_CIDR    uint8    `json:"src-cidr"`    //bgp announce prefix of cidr.
	DST_IP string
	DST_CIDR    uint8
	SRC_PORT 
}

func Env_load() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")

	}
}

func main() {

	a := make(chan Announce)

	go JgobServer(a)

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="MY REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {
			w.Write([]byte("API is up and running\n"))
		}
	})

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="MY REALM"`)
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

			var res Announce
			var resOriginBufUint8 uint8

			if jsonBody["origin"] == "igp" {
				resOriginBufUint8 = 0
			} else if jsonBody["origin"] == "egp" {
				resOriginBufUint8 = 1
			} else if jsonBody["origin"] == "incomplete" {
				resOriginBufUint8 = 2
			} else {
				resOriginBufUint8 = 2
			}
			res.ORIGIN = resOriginBufUint8

			res.NEXTHOP = jsonBody["nexthop"]

			if strings.Contains(jsonBody["aspath"], " ") {
				resAspathBufAry := strings.Split(jsonBody["aspath"], " ")
				for _, s := range resAspathBufAry {
					sInt, err := strconv.Atoi(s)
					if err != nil {
						fmt.Printf("error5: %s\n", err)
					}
					res.ASPATH = append(res.ASPATH, uint32(sInt))
				}
			} else {
				resAspathBufInt, err := strconv.Atoi(jsonBody["aspath"])
				if err != nil {
					fmt.Printf("error6: %s\n", err)
				}
				res.ASPATH = append(res.ASPATH, uint32(resAspathBufInt))
			}

			res.ADDRESS = jsonBody["address"]

			resCidrBufInt, err := strconv.Atoi(jsonBody["cidr"])
			if err != nil {
				fmt.Printf("error1: %s\n", err)
			}
			res.CIDR = uint8(resCidrBufInt)

			a <- res

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

