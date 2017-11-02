# jgob
Rest HTTPS API with json from GoBGP using bgp4 IPv4 flowspec [RFC5575](https://tools.ietf.org/html/rfc5575) daemon.  
Add flowspec route, throw json, receive json!!:kissing_heart:

## Overview
this code is under implement suite.
- [GoBGP](https://github.com/osrg/gobgp) (Using GoBGP as Golang Library)
- REST HTTPS API(having a unique URI return bgp infomation with json format)
- HTTPS Access log(using [go-http-logger](github.com/ajays20078/go-http-logger))
- Hooking syslog(using [logrus](https://github.com/sirupsen/logrus))
- Easy [Toml](github.com/BurntSushi/toml) config files.
- Having permanent routing table with json format.
- When Reloading processes, loading last install routes.
