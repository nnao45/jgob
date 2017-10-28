package main

import (
	//	"bufio"
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
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"google.golang.org/grpc"
	"io/ioutil"
	"log/syslog"
	"net/url"
	"os"
	//"strconv"
	"strings"
	"time"
)

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func cat(filename string) string {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return string(f)
}

func dog(text string, filename string) {
	if !exists(filename) {
		//os.MkdirAll(GOBGPHOME, 0600)
		os.Create(filename)
	}

	data := []byte(text)
	err := ioutil.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		log.Error("Unable to loading, ", filename)
	}
}

func JgobServer(achan, schan, rchan chan string) {
	Env_load()

	//log.SetLevel(log.DebugLevel)
	//gobgpdLogFile, err := os.OpenFile("gobgpd.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if err != nil {
	//	panic(err)
	//}
	//log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: true})
	//log.SetOutput(gobgpdLogFile)
	log.SetOutput(ioutil.Discard)

	if err := addSyslogHook(":syslog", "syslog"); err != nil {
		log.Error("Unable to connect to syslog daemon, ", "syslog")
	}

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
		log.Error(err)
	}

	n = &config.Neighbor{
		Config: config.NeighborConfig{
			NeighborAddress: "10.14.10.17",
			PeerAs:          65501,
			PeerType:        config.PEER_TYPE_INTERNAL,
		},
	}

	if err := s.AddNeighbor(n); err != nil {
		log.Error(err)
	}

	lock := make(chan struct{}, 0)
	go func() {
		<-lock
		log.Info("Starting Check the HTTP API...")
		x := 0
		for {
			if x > 2 {
				log.Fatal("oh,sorry, unable to access http api...")
				os.Exit(1)
			}
			if curlCheck(os.Getenv("USERNAME"), os.Getenv("PASSWORD")) {
				log.Info("OK, Access the HTTP API.")
				break
			} else {
				time.Sleep(500 * time.Millisecond)
				x++
			}
		}

		log.Info("Starting installing the routing table...")
		values := url.Values{}
		err := curlPost(values, cat("jgob.route"), os.Getenv("USERNAME"), os.Getenv("PASSWORD"))
		if err != nil {
			log.Error("Unable to loading route's json")
		}
		log.Info("Finish the installing Jgob's routing table.")
	}()

	timeout := grpc.WithTimeout(time.Second)
	conn, rpcErr := grpc.Dial("localhost:50051", timeout, grpc.WithBlock(), grpc.WithInsecure())
	if rpcErr != nil {
		log.Fatal("GoBGP is probably not running on the local server ... Please start gobgpd process !\n")
		log.Fatal(rpcErr)
		return
	}

	var count int
	for {
		select {
		case c := <-achan:
			client := api.NewGobgpApiClient(conn)
			var err error
			if strings.Contains(c, "match") {
				_, err = pushNewFlowSpecPath(client, c, "IPv4")
				log.Info("Adding flowspec prefix is ", c)
			} else {
				err = deleteFlowSpecPath(client, c)
				log.Info("Deleting flowspec uuid , ", c)
			}
			if err != nil {
				log.Error(err)
			}
			dog(showFlowSpecRib(client), "jgob.route")
		case req := <-schan:
			switch req {
			case "route":
				client := api.NewGobgpApiClient(conn)
				rchan <- showFlowSpecRib(client)
			case "nei":
				client := api.NewGobgpApiClient(conn)
				//var rsum string
				//for _, s := range showBgpNeighbor(client) {
				//	rsum = rsum + s
				//}
				//rchan <- rsum
				rchan <- showBgpNeighbor(client)
			}
		default:
			if count == 0 {
				count++
				lock <- struct{}{}
			}
		}
	}
}

func addSyslogHook(host, facility string) error {
	dst := strings.SplitN(host, ":", 2)
	network := ""
	addr := ""
	if len(dst) == 2 {
		network = dst[0]
		addr = dst[1]
	}

	priority := syslog.Priority(0)
	switch facility {
	case "kern":
		priority = syslog.LOG_KERN
	case "user":
		priority = syslog.LOG_USER
	case "mail":
		priority = syslog.LOG_MAIL
	case "daemon":
		priority = syslog.LOG_DAEMON
	case "auth":
		priority = syslog.LOG_AUTH
	case "syslog":
		priority = syslog.LOG_SYSLOG
	case "lpr":
		priority = syslog.LOG_LPR
	case "news":
		priority = syslog.LOG_NEWS
	case "uucp":
		priority = syslog.LOG_UUCP
	case "cron":
		priority = syslog.LOG_CRON
	case "authpriv":
		priority = syslog.LOG_AUTHPRIV
	case "ftp":
		priority = syslog.LOG_FTP
	case "local0":
		priority = syslog.LOG_LOCAL0
	case "local1":
		priority = syslog.LOG_LOCAL1
	case "local2":
		priority = syslog.LOG_LOCAL2
	case "local3":
		priority = syslog.LOG_LOCAL3
	case "local4":
		priority = syslog.LOG_LOCAL4
	case "local5":
		priority = syslog.LOG_LOCAL5
	case "local6":
		priority = syslog.LOG_LOCAL6
	case "local7":
		priority = syslog.LOG_LOCAL7
	}

	hook, err := lSyslog.NewSyslogHook(network, addr, syslog.LOG_INFO|priority, "gobgpd")
	if err != nil {
		return err
	}
	log.AddHook(hook)
	return nil
}

func pushNewFlowSpecPath(client api.GobgpApiClient, myCommand string, myAddrFam string) ([]byte, error) {
	if myAddrFam == "IPv4" {
		path, err := cmd.ParsePath(bgp.RF_FS_IPv4_UC, strings.Split(myCommand, " "))
		if err != nil {
			log.Fatal(err)
		}
		return (addFlowSpecPath(client, []*table.Path{path}))
	}
	if myAddrFam == "IPv6" {
		path, err := cmd.ParsePath(bgp.RF_FS_IPv6_UC, strings.Split(myCommand, " "))
		if err != nil {
			log.Fatal(err)
		}
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
		log.Error("Something gone wrong with UUID converion into bytes: %s\n", err)
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

	wc := len(myNativeTable.GetSortedDestinations())

	for i, d := range myNativeTable.GetSortedDestinations() {
		var ps []*table.Path
		ps = d.GetAllKnownPathList()
		s := showRouteToItem(ps)
		sum = sum + s
		if i+1 < wc {
			sum = sum + ","
		}
	}
	sum = "[" + sum + "]"
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

		apStr = `"aspath":"` + apStr + `"`

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
				if strings.Contains(s, "no-export") {
					s = strings.Replace(s, " no-export", "no-export", 1)
				}
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
				if strings.Contains(s, "protocol") {
					s = strings.Replace(s, "protocol:==", `protocol:`, 1)
					s = strings.Trim(s, " ")
				}
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

		uuid := `"uuid":"` + p.UUID().String() + `",`

		// fill up the tree with items
		str := fmt.Sprintf("{%s %s \"attrs\":{%s %s %s}}", uuid, age, nlriStr, attrStr, apStr)
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

func showBgpNeighbor(client api.GobgpApiClient) string {
	var dumpResult string
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
	format_fsm := func(admin api.PeerState_AdminState, fsm string) string {
		switch admin {
		case api.PeerState_DOWN:
			return "Idle(Admin)"
		case api.PeerState_PFX_CT:
			return "Idle(PfxCt)"
		}
		return fsm
	}

	for i, p := range m {
		neigh := p.Conf.NeighborAddress
		if p.Conf.NeighborInterface != "" {
			neigh = p.Conf.NeighborInterface
		}
		peer := `"peer":"` + fmt.Sprint(neigh) + `"`
		age := `"age":"` + fmt.Sprint(timedelta[i]) + `"`
		state := `"state":"` + format_fsm(p.Info.AdminState, p.Info.BgpState) + `"`
		as := `"as":"` + fmt.Sprint(p.Conf.PeerAs) + `"`
		advertised := `"advertised":"` + fmt.Sprint(p.Info.Advertised) + `"`
		received := `"received":"` + fmt.Sprint(p.Info.Received) + `"`
		accepted := `"accepted":"` + fmt.Sprint(p.Info.Accepted) + `"`

		dumpResult = dumpResult + fmt.Sprintf("{%s, %s, %s, \"attrs\":{%s, \"routes\":{%s, %s, %s}}}", peer, age, state, as, advertised, received, accepted)
		if i+1 < len(m) {
			dumpResult = dumpResult + ","
		}
	}
	dumpResult = "[" + dumpResult + "]"
	return dumpResult
}
