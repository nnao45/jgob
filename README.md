# jgob
Rest HTTPS API with json from GoBGP using bgp4 IPv4 flowspec [RFC5575](https://tools.ietf.org/html/rfc5575) daemon.  

## Motivation
I want to make very Mutual cooperation & HTTP frendly BGP daemon.  
So this daemon, When You add flowspec route, throw json, receive json!!:kissing_heart:

## Overview
this code is under implement suite.
- [GoBGP](https://github.com/osrg/gobgp) (Using GoBGP as Golang Library)
- REST HTTPS API(using [mux](github.com/gorilla/mux), having a unique URI return bgp infomation with json format)
- HTTPS Access log(using [go-http-logger](github.com/ajays20078/go-http-logger))
- Hooking syslog(using [logrus](https://github.com/sirupsen/logrus))
- Easy [Toml](github.com/BurntSushi/toml) config files.
- Having permanent routing table with json format.
- When Reloading processes, loading last install routes.

Running gRPC server with this Gobgp daemon,  
so you want to use "gobgp" client command, you will.

## HTTPS API Map

```bash

/--┐
   |---/global ... show global configuration of Running Gobgp. 
   |
   |---/nei ...... show bgp ipv4 flowspec neighbor of Running Gobgp. 
   |
   |---/route ...  show rib of address-family ipv4 flowspec of Running Gobgp. 
   |
   |---/add ...... adding ipv4 flowspec routes with more bgp attribute.
   |
   |---/del ...... deleting ipv4 flowspec routes from uuid.
   |
   |---/reload ... reloading rib from jgob.route(it's danger API...)
```

## jgob Have json fomat routing table
Plain text.
```bash
root@ubu-bgp:/godev/jgob/jgob_IPv4Flowspec# cat jgob.route | jq .
[
  {
    "attrs": {
      "destination": "3.0.0.0/24",
      "source": "2.0.0.0/24",
      "protocol": "udp",
      "destination-port": " ==22",
      "source-port": " ==80",
      "origin": " e",
      "extcomms": "2000.000000",
      "aspath": ""
    }
  },
  {
    "attrs": {
      "destination": "33.0.0.0/24",
      "source": "22.0.0.0/24",
      "protocol": "udp",
      "destination-port": " ==22",
      "source-port": " ==80",
      "origin": " e",
      "extcomms": "2000.000000",
      "aspath": ""
    }
  },
  {
    "attrs": {
      "destination": "93.0.0.0/24",
      "source": "92.0.0.0/24",
      "protocol": "udp",
      "destination-port": " ==22",
      "source-port": " ==80",
      "origin": " e",
      "extcomms": "2000.000000",
      "aspath": ""
    }
  },
  {
    "attrs": {
      "destination": "192.168.0.0/24",
      "source": "10.0.0.0/24",
      "protocol": "tcp",
      "destination-port": " ==9999",
      "source-port": " ==22222",
      "origin": " i",
      "extcomms": "100000.000000",
      "aspath": "65500,65000"
    }
  }
]
```

## Demo
### infra
```bash

[jgob#1(10.0.0.1)]=====[jgob#2(10.0.0.2)]

```
### config
```bash

$ jgob1
[jgobconfig]
as = 65501
router-id = "10.0.0.1"

[[jgobconfig.neighbor-config]]
peer-as = 65501
neighbor-address = "10.0.0.2"
peer-type = "internal"

$ jgob2
[jgobconfig]
as = 65501
router-id = "10.0.0.2"

[[jgobconfig.neighbor-config]]
peer-as = 65501
neighbor-address = "10.0.0.1"
peer-type = "internal"

```


### Show Bgp config status
#### show bgp neighbor
GET "/nei"
![result](https://github.com/nnao45/naoGifRepo/blob/master/showneijpg.jpg)
#### show route flowspec
GET "/route"
![result](https://github.com/nnao45/naoGifRepo/blob/master/showroute.jpg)
### Add Bgp route
POST new routes to "/add"
![result](https://github.com/nnao45/naoGifRepo/blob/master/post_newroute.jpg)  
Done, And received "uuid(it's example, "74a0a6c7-d28d-484f-a168-055014cbdba1")".  
this is adding route's universally unique id. This Using Deleting & Rib Management.
![result](https://github.com/nnao45/naoGifRepo/blob/master/responsenewroute.jpg)  
GET "/route", You can find that uuid, "74a0a6c7-d28d-484f-a168-055014cbdba1"  
![result](https://github.com/nnao45/naoGifRepo/blob/master/lockroutes.jpg)
### Delete Bgp route
If you want to route delete, it's very easy.  
POST "/del" a route having uuid(if you will want to check uuid, GET "/route").
![result](https://github.com/nnao45/naoGifRepo/blob/master/delete.jpg)  
And receiving system messages.
![result](https://github.com/nnao45/naoGifRepo/blob/master/successer.jpg)  
(if server internal faild, msg's values in direct error messages)

## Info
- jgob can receving protocol "tcp", "udp", "icmp".
- jgob can receving flowsepc action (MBGP EXT_COMMUNITIES) "accept", "discard", "rate-limit".  
  this three action, using same keys "extcomms"

Why selecting args?? sorry, when jgob pasing json all gobgp option, json formating is very difficult.
You want to other option, you rewirte code, or make issue or pull request for me :)

## Writer & License
jgob was writed by nnao45 (WORK:Network Engineer, Twitter:@A_Resas, MAIL:n4sekai5y@gmail.com).  
This software is released under the MIT License, see LICENSE.
