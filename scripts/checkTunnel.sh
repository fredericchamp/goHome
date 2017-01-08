#!/bin/bash
#
# Look if tunnel is open 
# $1 must identify ssh process (i.e. "port:host:hostport" )
#

COUNT=`ps -ef | grep ssh | grep $1 | wc | cut -c6-7`
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo "ps | grep ... fail"
	exit 1
fi

if [ ${COUNT} != 0 ]
then
	# SSH Tunnel found for $1
	echo "1"
else
	# No SSH Tunnel found for $1
	echo "0"
fi

exit 0
