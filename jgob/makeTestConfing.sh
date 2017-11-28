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
