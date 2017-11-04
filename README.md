[![Travis CI](https://travis-ci.org/nnao45/jgob.svg?branch=master)](https://travis-ci.org/nnao45/jgob)
[![Go Report Card](https://goreportcard.com/badge/github.com/nnao45/jgob)](https://goreportcard.com/report/github.com/nnao45/jgob)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/nnao45/jgob/master/LICENSE)
# jgob
Rest HTTPS API with json from [GoBGP](https://github.com/osrg/gobgp) using bgp4 IPv4 flowspec [RFC5575](https://tools.ietf.org/html/rfc5575) daemon.  

## Motivation
Concept, "Show config & Announce BGP UPDATE, throw json, receive json":kissing_heart:  
I want to make very Mutual cooperation & very HTTP frendly & very very simple flowspec BGP daemon.:laughing:  
## Overview
this code is under implement suite.
- [GoBGP](https://github.com/osrg/gobgp) (Using GoBGP as Golang Library, so jgob get values from Native GoBGP API return)
- REST HTTPS API(using [mux](https://github.com/gorilla/mux), having a unique URI return bgp infomation with json format)
- HTTPS Access log(using [go-http-logger](https://github.com/ajays20078/go-http-logger))
- Hooking syslog(using [logrus](https://github.com/sirupsen/logrus))
- Easy [Toml](https://github.com/BurntSushi/toml) config files.
- Having permanent routing table with json format.
- When Reloading processes, loading last install routes.
- Can remark route put a string which you like in a "remark" field. 

Running gRPC server with this Gobgp daemon,  
so you want to use "gobgp" client command, you will.

## Usage
Let's build jgob
```bash
$ git clone https://github.com/nnao45/jgob
$ cd jgob
$ go build
```
jgob use SSL, so you must make certification object.
If you don't have, use `makeSSL.sh`.
```bash
$ cat makeSSL.sh
#!/bin/sh

openssl genrsa 2048 > myself.key
openssl req -new -key myself.key <<EOF > myself.csr
JP
Tokyo
Japari Town
Japari Company
Japari Section
nyanpasu.com


EOF
openssl x509 -days 3650 -req -signkey myself.key < myself.csr > myself.crt
mkdir -p ssl/development/
mv myself.crt ssl/development
mv myself.csr ssl/development
mv myself.key ssl/development
```
It's joke infomation :stuck_out_tongue_winking_eye:
Do Use only to test.

And, jgob's Usage...

```bash
Usage:
    jgob_IPv4Flowspec [-r route-file] [-f config-file]

Examples:
    jgob_IPv4Flowspec
    jgob_IPv4Flowspec -r test.rib -f tokyo.tml
```
## HTTPS API Map

```bash

/--┐
   |---/test ..... simple test URI. Check living HTTPS API.
   |
   |---/global ... show global configuration of Running Gobgp. 
   |
   |---/nei ...... show bgp ipv4 flowspec neighbor of Running Gobgp. 
   |
   |---/route .... show a rib of address-family ipv4 flowspec of Running Gobgp. 
   |
   |---/remark ... show route's remark and uuid in a rib. 
   |
   |---/add ...... adding ipv4 flowspec routes with more bgp attribute.
   |
   |---/del ...... deleting ipv4 flowspec routes from uuid.
   |
   |---/reload ... reloading rib from jgob.route(it's danger API...)
```

## jgob have json fomat routing table
Plain text.
```bash
root@ubu-bgp:/godev/jgob/jgob_IPv4Flowspec# cat jgob.route | jq .
[
  {
    "remark":"hoge"
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
    "remark":"piyo"
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
    "remark":"huga"
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
    "remark":"ponyo"
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
It's so unique? :kissing_smiling_eyes:

## jgob's json struct
```bash
type Prefix struct {
        Remark  string `json:"remark"`  //remarking this route, it's filed you take it easy to write. 
        Uuid    string `json:"uuid"`    //this route's universally unique id.
        Age     string `json:"age"`     //this route's aging time.
        Attrs struct {
                Aspath      string `json:"aspath"`              //this route flowspec attribute's as path.
                Protocol    string `json:"protocol"`            //this route flowspec attribute's protobcol.
                Src         string `json:"source"`              //this route flowspec attribute's src address.
                Dst         string `json:"destination"`         //this route flowspec attribute's dst address.
                SrcPort     string `json:"source-port"`         //this route flowspec attribute's src port.
                DstPort     string `json:"destination-port"`    //this route flowspec attribute's dst port.  
                Origin      string `json:"origin"`              //this route flowspec attribute's origin.  
                Communities string `json:"community"`           //this route flowspec attribute's community. 
                Extcomms    string `json:"extcomms"`            //this route flowspec attribute's extra community(for example, accept, discard, or rate-limit bps value).
        }
}

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
jgob config is very simple.
```bash
[jgobconfig]
as = <local-as>
router-id = <router-id>

[[jgobconfig.neighbor-config]]
peer-as = <remote-as>
neighbor-address = <neighbor-address>
peer-type = <peer-type>
```
address-family fixed, ipv4-flowspec.
You must use only these param, and toml format.

### Show Bgp config & status
#### show bgp neighbor
GET "/nei"
![result](https://github.com/nnao45/naoGifRepo/blob/master/showneijpg.jpg)
#### show route flowspec
GET "/route"
![result](https://github.com/nnao45/naoGifRepo/blob/master/shoshoroute.jpg)
### Add Bgp route
POST new routes to "/add" (multipath is ok, adding in array :innocent:)  
Don't need to "age" value, "uuid" value.
![result](https://github.com/nnao45/naoGifRepo/blob/master/addroute.jpg)  
Done, And received "uuid"(adding route's universally unique id), and "remark"(adding route's remark, free string)  
This is Used Deleting & Rib Management.
![result](https://github.com/nnao45/naoGifRepo/blob/master/responsenwroute.jpg)  
### Delete Bgp route
If you want to route delete, it's very easy.(also, multipath is ok, adding in array :innocent:)  
POST "/del" a route having uuid(if you will want to check uuid, GET "/route").  
Need to only "uuid" value.
![result](https://github.com/nnao45/naoGifRepo/blob/master/deleteroute.jpg)  
And receiving delete route's uuid, remark, and system messages.
![result](https://github.com/nnao45/naoGifRepo/blob/master/deletesuccess.jpg)  
(if server internal faild, system messsages will be values in direct error messages)

## Info
- I think that jgob is as flowspec controller, so may not be received routes.
- jgob is running auto sync interval 1sec "jgob.route" and GoBGP Rib(If you use "gobgp" cmd, no problem).
- jgob's global configuration, Intentionally can't change(add neighbor, delete neighbor, change router-id...), but you can use "gobgp" cmd, so this operation, use cmd.
- jgob can receving protocol "tcp", "udp", "icmp".
- jgob can receving flowsepc action (MBGP EXT_COMMUNITIES) "accept", "discard", "rate-limit".  
  this three action, using same keys "extcomms"

Why selecting args?? sorry, when jgob pasing json all gobgp option, json formating is very difficult.
You want to other option, you rewirte code, or make issue or pull request for me :)

## Release note
- now, βversion, may not stable:sweat_smile:

***Have a nice go hacking days***:sparkles::wink:

## Writer & License
jgob was writed by nnao45 (WORK:Network Engineer, Twitter:@A_Resas, MAIL:n4sekai5y@gmail.com).  
This software is released under the MIT License, see LICENSE.
