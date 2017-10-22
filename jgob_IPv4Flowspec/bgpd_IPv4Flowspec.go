package main

import (
	"strings"
	"context"
	"fmt"
	api "github.com/osrg/gobgp/api"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/packet/bgp"
	gobgp "github.com/osrg/gobgp/server"
	"github.com/osrg/gobgp/table"
	log "github.com/sirupsen/logrus"
	"time"
	"google.golang.org/grpc"
	"github.com/osrg/gobgp/gobgp/cmd"
)

func JgobServer(achan chan string, schan chan struct{}, rchan chan string) {
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
			As:       65000,
			RouterId: "10.0.255.254",
			Port:     -1, // gobgp won't listen on tcp:179
		},
	}

	if err := s.Start(global); err != nil {
		log.Fatal(err)
	}

	// neighbor configuration
	n := &config.Neighbor{
		Config: config.NeighborConfig{
			NeighborAddress: "10.0.255.1",
			PeerAs:          65000,
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
			_, err := PushNewFlowSpecPath(client, c, "IPv4")
			if err != nil{
				panic(err)
			}
		case <- schan:
			client := api.NewGobgpApiClient(conn)
			rchan <- ShowFlowSpecRib(client)
		}
	}
}


func PushNewFlowSpecPath(client api.GobgpApiClient, myCommand string, myAddrFam string) ([]byte, error) {
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

func ShowFlowSpecRib(client api.GobgpApiClient) string{
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

func showRouteToItem(pathList []*table.Path) string{
	maxPrefixLen := 100
	maxNexthopLen := 20
	var sum string

	now := time.Now()
	for _, p := range pathList {
		nexthop := "fictitious"
		if n := p.GetNexthop(); n != nil {
			nexthop = p.GetNexthop().String()
		}

		s := make([]string, 0, 5)
		aspath := make([]string, 0, 5)
		aspath = append(aspath , "AsPath:")
		for _, a := range p.GetPathAttrs() {
			switch a.GetType() {
			case bgp.BGP_ATTR_TYPE_NEXT_HOP, bgp.BGP_ATTR_TYPE_MP_REACH_NLRI:
				continue
			case bgp.BGP_ATTR_TYPE_AS_PATH, bgp.BGP_ATTR_TYPE_AS4_PATH:
				aspath = append(aspath, a.String())
			default:
				s = append(s, a.String())
			}
		}
		pattrstr := fmt.Sprint(s)

		if maxNexthopLen < len(nexthop) {
			maxNexthopLen = len(nexthop)
		}

		nlri := p.GetNlri()

		if maxPrefixLen < len(nlri.String()) {
			maxPrefixLen = len(nlri.String())
		}

		nexthop = "[Nexthop:" + nexthop + "]"

		age := formatTimedelta(int64(now.Sub(p.GetTimestamp()).Seconds()))

		age = "[Age:" + age + "]"

		uuid := "[UUID:" + p.UUID().String() + "]"

		// fill up the tree with items
		str :=  fmt.Sprintf("%s %s %s %s %s %s\n", nlri.String(), nexthop, aspath, age, pattrstr, uuid)
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
