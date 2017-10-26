#!/bin/sh

openssl genrsa 2048 > myself.key
openssl req -new -key myself.key > myself.csr
openssl x509 -days 3650 -req -signkey myself.key < myself.csr > myself.crt
mkdir -p ssl/development/
mv myself.crt ssl/development
mv myself.csr ssl/development
mv myself.key ssl/development
