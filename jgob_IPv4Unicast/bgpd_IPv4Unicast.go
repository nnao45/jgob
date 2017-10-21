package main

import (
	"fmt"
	api "github.com/osrg/gobgp/api"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/packet/bgp"
	gobgp "github.com/osrg/gobgp/server"
	"github.com/osrg/gobgp/table"
	log "github.com/sirupsen/logrus"
	"time"
)

func JgobServer(receivedCmd chan Announce) {
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
			PeerAs:          65001,
		},
	}

	if err := s.AddNeighbor(n); err != nil {
		log.Fatal(err)
	}

	// add routes(example)
	/*
	attrs := []bgp.PathAttributeInterface{
		bgp.NewPathAttributeOrigin(0),
		bgp.NewPathAttributeNextHop("10.0.255.254"),
		bgp.NewPathAttributeAsPath([]bgp.AsPathParamInterface{bgp.NewAs4PathParam(bgp.BGP_ASPATH_ATTR_TYPE_SEQ, []uint32{4000, 400000, 300000, 40001})}),
	}
	if _, err := s.AddPath("", []*table.Path{table.NewPath(nil, bgp.NewIPAddrPrefix(24, "10.0.0.0"), false, attrs, time.Now(), false)}); err != nil {
		log.Fatal(err)
	}
	*/

	// monitor new routes
	w := s.Watch(gobgp.WatchBestPath(false))
	var r Announce
	for {
		select {
		case ev := <-w.Event():
			switch msg := ev.(type) {
			case *gobgp.WatchEventBestPath:
				for _, path := range msg.PathList {
					// do something useful
					fmt.Println(path)
				}
			}
			case r = <-receivedCmd:
			attrs := []bgp.PathAttributeInterface{
				bgp.NewPathAttributeOrigin(r.ORIGIN),
				bgp.NewPathAttributeNextHop(r.NEXTHOP),
				bgp.NewPathAttributeCommunities(r.COMMUNITIES),
				bgp.NewPathAttributeAsPath([]bgp.AsPathParamInterface{bgp.NewAs4PathParam(bgp.BGP_ASPATH_ATTR_TYPE_SEQ, r.ASPATH)}),
			}
			if _, err := s.AddPath("", []*table.Path{table.NewPath(nil, bgp.NewIPAddrPrefix(r.CIDR, r.ADDRESS), false, attrs, time.Now(), false)}); err != nil {
				log.Fatal(err)
			}
		}
	}
}
