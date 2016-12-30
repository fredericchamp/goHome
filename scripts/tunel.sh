#!/bin/bash
#

# First check if tunnel is not already open
COUNT=`ps -ef | grep ssh | grep $2 | wc | cut -c6-7`
if [ ${COUNT} != 0 ]
then
	echo "SSH Tunel is up"
	exit 0
fi

# Open a SSH Tunel 
/usr/bin/ssh $@
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo "SSH Tunel fail"
	exit 1
fi

echo "SSH Tunel ok"
exit 0

