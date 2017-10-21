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

func JgobServer(cmdLine chan string) {
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
		case c := <-cmdLine:
			client := api.NewGobgpApiClient(conn)
			_, err := PushNewFlowSpecPath(client, c, "IPv4")
			if err != nil{
				panic(err)
			}
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
