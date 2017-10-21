package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
//	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Announce struct {
	ORIGIN      uint8    `json:"origin"`      //bgp origin-as attr.
	NEXTHOP     string   `json:"nexthop"`     //bgp nexthop attr.
	ASPATH      []uint32 `json:"aspath"`      //bgp as-path attr.
	COMMUNITIES []uint32 `json:"communities"` //bgp communities attr.
	ADDRESS     string   `json:"address"`     //bgp announce prefix of address.
	CIDR        uint8    `json:"cidr"`        //bgp announce prefix of cidr.
}

const (
	COMMUNITY_INTERNET					= 0x00000000
	COMMUNITY_PLANNED_SHUT                                  = 0xffff0000
	COMMUNITY_ACCEPT_OWN                                    = 0xffff0001
	COMMUNITY_ROUTE_FILTER_TRANSLATED_v4                    = 0xffff0002
	COMMUNITY_ROUTE_FILTER_v4                               = 0xffff0003
	COMMUNITY_ROUTE_FILTER_TRANSLATED_v6                    = 0xffff0004
	COMMUNITY_ROUTE_FILTER_v6                               = 0xffff0005
	COMMUNITY_LLGR_STALE                                    = 0xffff0006
	COMMUNITY_NO_LLGR                                       = 0xffff0007
	COMMUNITY_BLACKHOLE                                     = 0xffff029a
	COMMUNITY_NO_EXPORT                                     = 0xffffff01
	COMMUNITY_NO_ADVERTISE                                  = 0xffffff02
	COMMUNITY_NO_EXPORT_SUBCONFED                           = 0xffffff03
	COMMUNITY_NO_PEER                                       = 0xffffff04
)

var WellKnownCommunityNameMap = map[string]uint32{
	"internet":                   COMMUNITY_INTERNET,
	"planned-shut":               COMMUNITY_PLANNED_SHUT,
	"accept-own":                 COMMUNITY_ACCEPT_OWN,
	"route-filter-translated-v4": COMMUNITY_ROUTE_FILTER_TRANSLATED_v4,
	"route-filter-v4":            COMMUNITY_ROUTE_FILTER_v4,
	"route-filter-translated-v6": COMMUNITY_ROUTE_FILTER_TRANSLATED_v6,
	"route-filter-v6":            COMMUNITY_ROUTE_FILTER_v6,
	"llgr-stale":                 COMMUNITY_LLGR_STALE,
	"no-llgr":                    COMMUNITY_NO_LLGR,
	"blackhole":                  COMMUNITY_BLACKHOLE,
	"no-export":                  COMMUNITY_NO_EXPORT,
	"no-advertise":               COMMUNITY_NO_ADVERTISE,
	"no-export-subconfed":        COMMUNITY_NO_EXPORT_SUBCONFED,
	"no-peer":                    COMMUNITY_NO_PEER,
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
			w.Header().Set("WWW-Authenticate", `Basic realm="JGOB REALM"`)
			w.WriteHeader(401)
			w.Write([]byte("401 Unauthorized\n"))
			return
		} else {
			w.Write([]byte("JGOB is up and running\n"))
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

			var res Announce

			if jsonBody["origin"] == "igp" {
				res.ORIGIN = 0
			} else if jsonBody["origin"] == "egp" {
				res.ORIGIN = 1
			} else if jsonBody["origin"] == "incomplete" {
				res.ORIGIN = 2
			} else {
				res.ORIGIN = 2
			}

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

			if strings.Contains(jsonBody["communities"], " ") {
				resCommunitiesAry := strings.Split(jsonBody["communities"], " ")
				for _, s := range resCommunitiesAry {
					if n, ok := WellKnownCommunityNameMap[s]; ok {
						res.COMMUNITIES = append(res.COMMUNITIES, n)
					} else {
						sInt, err := strconv.Atoi(s)
						if err != nil {
							fmt.Printf("error6: %s\n", err)
						}
						res.COMMUNITIES = append(res.COMMUNITIES, 0xffff&uint32(sInt))
					}
				}
			} else {
				if n, ok := WellKnownCommunityNameMap[jsonBody["communities"]]; ok {
					res.COMMUNITIES = append(res.COMMUNITIES, n)
				} else {
					sInt, err := strconv.Atoi(jsonBody["communities"])
					if err != nil {
						fmt.Printf("error6: %s\n", err)
					}
					res.COMMUNITIES = append(res.COMMUNITIES, 0xffff&uint32(sInt))
				}
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

