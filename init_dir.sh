#!/bin/bash
#

VAR_DIR="/var/goHome"
SRC_DIR="${GOPATH}/src/github.com/fredericchamp/goHome"

sudo mkdir ${VAR_DIR}
sudo chmod 777  ${VAR_DIR}
cd ${VAR_DIR}

mkdir log


ln -s ${SRC_DIR}/certificats
ln -s ${SRC_DIR}/www

ln -s ${SRC_DIR}/init.sql
ln -s ${SRC_DIR}/perso.sql


