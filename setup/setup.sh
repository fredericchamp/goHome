#!/bin/bash
#

VAR_DIR="/var/goHome"
SRC_DIR="${GOPATH}/src/github.com/fredericchamp/goHome"

# Create base dir for goHome
if [ ! -d ${VAR_DIR} ]
then
	sudo mkdir ${VAR_DIR}
	sudo chmod 777  ${VAR_DIR}
fi

# Create log dir for goHome
if [ ! -d ${VAR_DIR}/log ]
then
	mkdir ${VAR_DIR}/log
fi

# Create backup dir for goHome
if [ ! -d ${VAR_DIR}/backup ]
then
	mkdir ${VAR_DIR}/backup
fi

# Create link for SSL certificats
if [ ! -d ${VAR_DIR}/certificats ]
then
	ln -s ${SRC_DIR}/certificats ${VAR_DIR}/certificats
fi

# Create link www document root
if [ ! -d ${VAR_DIR}/www ]
then
	ln -s ${SRC_DIR}/www ${VAR_DIR}/www
fi

# Create link for sql files
for i in `ls -1 ${SRC_DIR}/setup/*.sql`
do
	if [ ! -f ${VAR_DIR}/`basename $i` ]
	then
		ln -s $i ${VAR_DIR}/`basename $i`
	fi
done

# Create link for binary and shell scripts
if [ ! -f /usr/local/bin/goHome ]
then
	sudo ln -s ${GOPATH}/bin/goHome /usr/local/bin/goHome
fi

for i in `ls -1 ${SRC_DIR}/scripts/`
do
	if [ ! -f /usr/local/bin/$i ]
	then
		sudo ln -s ${SRC_DIR}/scripts/$i /usr/local/bin/$i
	fi
done

