#!/bin/sh

cat << EOT > config.tml
[jgobconfig]
as = 65501
router-id = "10.0.0.1"

[[jgobconfig.neighbor-config]]
peer-as = 65501
neighbor-address = "10.0.0.2"
peer-type = "internal"
EOT

cat << EOT > .env
USERNAME=user
PASSWORD=pass
EOT

cat << EOT > jgob.route
[{"remark":"kiiiiiiiiiiiiin!!!!!","attrs":{"destination":"111.0.0.0/24","source":"222.0.0.0/24","protocol":"udp","destination-port":" ==22222","source-port":" ==1111","origin":"e","extcomms":"99999.000000","aspath":"1,2,3,4,5,6,7"}}]
EOT
