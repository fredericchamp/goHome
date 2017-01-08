#!/bin/bash
#
# Open SSH Tunnel (after killing existing tunnel for same "port:host:hostport" )
#

# Close existing SSH tunnel for $2 if any $2 is "port:host:hostport"
SSH_PID=`ps -ef | grep ssh | grep $2 | cut -c9-15`
if [ "${SSH_PID}" != "" ]
then
	/bin/kill ${SSH_PID}
	RESULT=$?
	if [ ${RESULT} != 0 ]
	then
		echo "kill ${SSH_PID} fail"
		exit 1
	fi
fi

# Open SSH Tunnel 
/usr/bin/ssh $@
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo "SSH Tunnel fail"
	exit 1
fi

echo "SSH Tunnel ok"
exit 0
