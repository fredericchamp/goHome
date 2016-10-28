#!/bin/bash
#
set -x

# ####################################################
# IMPORTANT First create a self signed CA
#
# mkdir certs crl newcerts private
# touch index.txt
# openssl req -new -newkey rsa:4096 -keyout private/goHomeCAkey.pem -out goHomeCAreq.pem -config ./goHome.cnf
# openssl ca -create_serial -out goHomeCAcert.pem -days 3650 -keyfile private/goHomeCAkey.pem -selfsign -extensions v3_ca -config ./goHome.cnf -infiles goHomeCAreq.pem
#

function CreateCA {
        echo "CA key file not found : need to create CA first !"
        echo ""
        echo "      mkdir certs crl newcerts private"
        echo "      touch index.txt"
        echo "      openssl req -new -newkey rsa:4096 -keyout private/goHomeCAkey.pem -out goHomeCAreq.pem -config ./goHome.cnf"
        echo "      openssl ca -create_serial -out goHomeCAcert.pem -days 3650 -keyfile private/goHomeCAkey.pem -selfsign -extensions v3_ca -config ./goHome.cnf -infiles goHomeCAreq.pem"
        echo ""
}

function Usage {
        echo "Usage : $0 <cert_base_name>"
        echo ""
        echo "      $0 john"
        echo ""
}

# ####################################################
# Check existence of CA key file
if [ ! -f ./private/goHomeCAkey.pem ]
then
        CreateCA
        exit 1
fi


# ####################################################
# Check running as root
#if [ `id -u` != 0 ]
#then
#        echo "$0 needs to be run as root"
#        exit 1
#fi

# ####################################################
# Check command line parameter : cert_base_name

if [ $# -lt 1 ]
then
        Usage
        exit 1
fi

name=$1

# ####################################################
# Creation .csr (demande de certificat)
openssl req -new -newkey rsa:2048 -keyout ${name}.key.pem -out ${name}.csr.pem -config goHome.cnf

# ####################################################
# Creation certificat
openssl ca -config goHome.cnf -out ./certs/${name}.crt.pem -infiles ${name}.csr.pem

# csr is now useless 
rm -f ${name}.csr.pem

# ####################################################
# Creation pkcs12
openssl pkcs12 -export -in ./certs/${name}.crt.pem -inkey ${name}.key.pem -out ${name}.p12 -name "Home (${name})"

# key should not be kept as it is securly (depending on the pass used) stored in p12 keystore
rm -f ${name}.key.pem
