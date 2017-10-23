package main

import (
	"bufio"
	"context"
	"fmt"
	api "github.com/osrg/gobgp/api"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/gobgp/cmd"
	"github.com/osrg/gobgp/packet/bgp"
	gobgp "github.com/osrg/gobgp/server"
	"github.com/osrg/gobgp/table"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"
)

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func dog(text string, filename string) {
	if !exists(filename) {
		//os.MkdirAll(GOBGPHOME, 0600)
		os.Create(filename)
	}

	data := []byte(text)
	err := ioutil.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
}

func JgobServer(achan, schan, rchan chan string) {
	Env_load()

	log.SetLevel(log.DebugLevel)
	s := gobgp.NewBgpServer()
	go s.Serve()

	// start grpc api server. this is not mandatory
	// but you will be able to use `gobgp` cmd with this.
	g := api.NewGrpcServer(s, ":50051")
	go g.Serve()

	// global configuration
	global := &config.Global{
		Config: config.GlobalConfig{
			As:       65501,
			RouterId: "172.30.1.176",
			Port:     -1, // gobgp won't listen on tcp:179
		},
	}

	if err := s.Start(global); err != nil {
		log.Fatal(err)
	}

	// neighbor configuration
	n := &config.Neighbor{
		Config: config.NeighborConfig{
			NeighborAddress: "10.14.10.16",
			PeerAs:          65501,
			PeerType:        config.PEER_TYPE_INTERNAL,
		},
		AfiSafis: []config.AfiSafi{
			config.AfiSafi{
				Config: config.AfiSafiConfig{
					AfiSafiName: "ipv4-flowspec",
					Enabled:     true,
				},
			},
		},
	}

	if err := s.AddNeighbor(n); err != nil {
		log.Fatal(err)
	}

	go func() {
		x := 0
		for {
			if x > 2 {
				os.Exit(1)
			}
			if curlCheck(os.Getenv("USERNAME"), os.Getenv("PASSWORD")) {
				break
			} else {
				time.Sleep(500 * time.Millisecond)
				x++
			}
		}
		last, err := os.Open("jgob.route")
		if err != nil {
			panic(err)
		}
		defer last.Close()
		lastscanner := bufio.NewScanner(last)
		for lastscanner.Scan() {
			route := lastscanner.Text()
			values := url.Values{}
			err := curlPost(values, route, os.Getenv("USERNAME"), os.Getenv("PASSWORD"))
			if err != nil {
				panic(err)
			}
		}

	}()

	timeout := grpc.WithTimeout(time.Second)
	conn, rpcErr := grpc.Dial("localhost:50051", timeout, grpc.WithBlock(), grpc.WithInsecure())
	if rpcErr != nil {
		fmt.Printf("GoBGP is probably not running on the local server ... Please start gobgpd process !\n")
		fmt.Println(rpcErr)
		return
	}

	for {
		select {
		case c := <-achan:
			client := api.NewGobgpApiClient(conn)
			if strings.Contains(c, "match") {
				_, err := pushNewFlowSpecPath(client, c, "IPv4")
				if err != nil {
					log.Fatal(err)
				}
			} else {
				err := deleteFlowSpecPath(client, c)
				if err != nil {
					log.Fatal(err)
				}
			}
			dog(showFlowSpecRib(client), "jgob.route")
		case req := <-schan:
			switch req {
			case "route":
				client := api.NewGobgpApiClient(conn)
				rchan <- showFlowSpecRib(client)
			case "nei":
				client := api.NewGobgpApiClient(conn)
				var rsum string
				for _, s := range showBgpNeighbor(client) {
					rsum = rsum + s
				}
				rchan <- rsum
			}
		}
	}
}

func pushNewFlowSpecPath(client api.GobgpApiClient, myCommand string, myAddrFam string) ([]byte, error) {
	if myAddrFam == "IPv4" {
		path, _ := cmd.ParsePath(bgp.RF_FS_IPv4_UC, strings.Split(myCommand, " "))
		return (addFlowSpecPath(client, []*table.Path{path}))
	}
	if myAddrFam == "IPv6" {
		path, _ := cmd.ParsePath(bgp.RF_FS_IPv6_UC, strings.Split(myCommand, " "))
		return (addFlowSpecPath(client, []*table.Path{path}))
	}
	return nil, nil
}

func addFlowSpecPath(client api.GobgpApiClient, pathList []*table.Path) ([]byte, error) {
	vrfID := ""
	resource := api.Resource_GLOBAL
	var uuid []byte
	for _, path := range pathList {
		r, err := client.AddPath(context.Background(), &api.AddPathRequest{
			Resource: resource,
			VrfId:    vrfID,
			Path:     api.ToPathApi(path),
		})
		if err != nil {
			return nil, err
		}
		uuid = r.Uuid
	}
	return uuid, nil
}

func deleteFlowSpecPath(client api.GobgpApiClient, myUuid string) error {
	byteUuid, err := uuid.FromString(myUuid)
	if err != nil {
		fmt.Printf("Something gone wrong with UUID converion into bytes: %s\n", err)
	}
	return deleteFlowSpecPathFromUuid(client, byteUuid.Bytes())
}

func deleteFlowSpecPathFromUuid(client api.GobgpApiClient, uuid []byte) error {
	var reqs []*api.DeletePathRequest
	var vrfID = ""
	resource := api.Resource_GLOBAL
	reqs = append(reqs, &api.DeletePathRequest{
		Resource: resource,
		VrfId:    vrfID,
		Uuid:     uuid,
		Family:   uint32(0),
	})
	for _, req := range reqs {
		if _, err := client.DeletePath(context.Background(), req); err != nil {
			return err
		}
	}
	return nil
}

func showFlowSpecRib(client api.GobgpApiClient) string {
	var dsts []*api.Destination
	var myNativeTable *table.Table
	var sum string
	resource := api.Resource_GLOBAL
	family, _ := bgp.GetRouteFamily("ipv4-flowspec")

	res, err := client.GetRib(context.Background(), &api.GetRibRequest{
		Table: &api.Table{
			Type:         resource,
			Family:       uint32(family),
			Name:         "",
			Destinations: dsts,
		},
	})
	if err != nil {
		return ""
	}
	myNativeTable, err = res.Table.ToNativeTable()

	for _, d := range myNativeTable.GetSortedDestinations() {
		var ps []*table.Path
		ps = d.GetAllKnownPathList()
		s := showRouteToItem(ps)
		sum = sum + s
	}
	return sum
}

func showRouteToItem(pathList []*table.Path) string {
	maxPrefixLen := 100
	maxNexthopLen := 20
	var sum string

	now := time.Now()
	for _, p := range pathList {
		nexthop := "fictitious"
		if n := p.GetNexthop(); n != nil {
			nexthop = p.GetNexthop().String()
		}

		attr := make([]string, 0, 5)
		aspath := make([]string, 0, 5)
		for _, a := range p.GetPathAttrs() {
			switch a.GetType() {
			case bgp.BGP_ATTR_TYPE_NEXT_HOP, bgp.BGP_ATTR_TYPE_MP_REACH_NLRI:
				continue
			case bgp.BGP_ATTR_TYPE_AS_PATH, bgp.BGP_ATTR_TYPE_AS4_PATH:
				aspath = append(aspath, a.String())
			default:
				attr = append(attr, a.String())
			}
		}

		apStr := strings.Replace(strings.Join(aspath, " "), " ", ",", -1)

		apStr = `"aspath":"` + apStr + `", `

		var attrStr string
		for _, s := range attr {
			s = strings.ToLower(s)
			if strings.Contains(s, "extcomms") && strings.Contains(s, "rate") {
				s = strings.Replace(s, "{extcomms: [rate: ", `"extcomms":"`, 1)
				s = strings.Replace(s, "]}", `"`, -1)
			} else if strings.Contains(s, "extcomms") && strings.Contains(s, "discard") {
				s = strings.Replace(s, "{extcomms: [", `"extcomms":"`, 1)
				s = strings.Replace(s, "]}", `"`, -1)
			} else {
				s = strings.Replace(s, ":", `":"`, -1)
				s = strings.Replace(s, "{", `"`, -1)
				s = strings.Replace(s, "}", `"`, -1)
			}
			attrStr = attrStr + s + ","
		}

		if maxNexthopLen < len(nexthop) {
			maxNexthopLen = len(nexthop)
		}

		nlri := p.GetNlri()
		var nlriAry []string
		var nlriStr string
		nlriAry = strings.Split(nlri.String(), "]")
		for _, s := range nlriAry {
			if s != "" {
				nlriStr = nlriStr + strings.Replace(s, "[", `"`, -1) + `", `
			}
		}
		nlriStr = strings.Replace(nlriStr, ":", `":"`, -1)

		if maxPrefixLen < len(nlri.String()) {
			maxPrefixLen = len(nlri.String())
		}

		//nexthop = "[Nexthop:" + nexthop + "]"

		age := formatTimedelta(int64(now.Sub(p.GetTimestamp()).Seconds()))

		age = `"age":"` + age + `",`

		uuid := `"uuid":"` + p.UUID().String() + `"`

		// fill up the tree with items
		str := fmt.Sprintf("{%s %s %s %s %s}\n", nlriStr, apStr, age, attrStr, uuid)
		sum = sum + str
	}
	return sum
}

func formatTimedelta(d int64) string {
	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}
	secs := u % 60
	u /= 60
	mins := u % 60
	u /= 60
	hours := u % 24
	days := u / 24

	if days == 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
	} else {
		return fmt.Sprintf("%dd ", days) + fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
	}
}

func showBgpNeighbor(client api.GobgpApiClient) []string {
	dumpResult := []string{}
	var NeighReq api.GetNeighborRequest
	NeighResp, e := client.GetNeighbor(context.Background(), &NeighReq)
	if e != nil {
		return dumpResult
	}
	m := NeighResp.Peers
	maxaddrlen := 0
	maxaslen := 0
	maxtimelen := len("Up/Down")
	timedelta := []string{}

	// sort.Sort(m)

	now := time.Now()
	for _, p := range m {
		if i := len(p.Conf.NeighborInterface); i > maxaddrlen {
			maxaddrlen = i
		} else if j := len(p.Conf.NeighborAddress); j > maxaddrlen {
			maxaddrlen = j
		}
		if len(fmt.Sprint(p.Conf.PeerAs)) > maxaslen {
			maxaslen = len(fmt.Sprint(p.Conf.PeerAs))
		}
		timeStr := "never"
		if p.Timers.State.Uptime != 0 {
			t := int64(p.Timers.State.Downtime)
			if p.Info.BgpState == "BGP_FSM_ESTABLISHED" {
				t = int64(p.Timers.State.Uptime)
			}
			timeStr = formatTimedelta(int64(now.Sub(time.Unix(int64(t), 0)).Seconds()))
		}
		if len(timeStr) > maxtimelen {
			maxtimelen = len(timeStr)
		}
		timedelta = append(timedelta, timeStr)
	}
	var format string
	format = "%-" + fmt.Sprint(maxaddrlen) + "s" + " %" + fmt.Sprint(maxaslen) + "s" + " %" + fmt.Sprint(maxtimelen) + "s"
	format += " %-11s |%11s %8s %8s\n"
	dumpResult = append(dumpResult, fmt.Sprintf(format, "Peer", "AS", "Up/Down", "State", "#Advertised", "Received", "Accepted"))
	format_fsm := func(admin api.PeerState_AdminState, fsm string) string {
		switch admin {
		case api.PeerState_DOWN:
			return "Idle(Admin)"
		case api.PeerState_PFX_CT:
			return "Idle(PfxCt)"
		}

		if fsm == "BGP_FSM_IDLE" {
			return "Idle"
		} else if fsm == "BGP_FSM_CONNECT" {
			return "Connect"
		} else if fsm == "BGP_FSM_ACTIVE" {
			return "Active"
		} else if fsm == "BGP_FSM_OPENSENT" {
			return "Sent"
		} else if fsm == "BGP_FSM_OPENCONFIRM" {
			return "Confirm"
		} else {
			return "Establ"
		}
	}

	for i, p := range m {
		neigh := p.Conf.NeighborAddress
		if p.Conf.NeighborInterface != "" {
			neigh = p.Conf.NeighborInterface
		}
		dumpResult = append(dumpResult, fmt.Sprintf(format, neigh, fmt.Sprint(p.Conf.PeerAs), timedelta[i], format_fsm(p.Info.AdminState, p.Info.BgpState), fmt.Sprint(p.Info.Advertised), fmt.Sprint(p.Info.Received), fmt.Sprint(p.Info.Accepted)))
	}
	return dumpResult
}
