#!/bin/bash
#
#set -x
#
#############################################
# Take a snap shot using IPwebcam install on an android device
# First call to http://<host:port>/ ... needed for some kine of "weakup for IPwebcam
# Then take the snapshot : http://<host:port>/photo.jpg
#
# $0 192.168.43.1:8080 destination_jpg_file

BASEDIR=/var/goHome
TMPFILE=/tmp/ipwebcam.jpg
DESTDIR=${BASEDIR}/www/capture
LOGFILE=${BASEDIR}/log/goHome.INFO

echo -e "`date +"I%m%d %H:%M:%S.000000"`\t00000\t$0 New IPwebcam snap shot for '$@'" | tee -a ${LOGFILE}

#############################################
# Check cmde line parameter
# expecting : <host:port> <dest file name> <convert parameters>
#

SRCHOST=$1
DESTFILE=${DESTDIR}/$2

if [ "${DESTFILE}" == "${DESTDIR}/" ]
then
	echo -e "`date +"E%m%d %H:%M:%S.000000"`\t00000\t$0 Missing parameters '$@'" | tee -a ${LOGFILE}
	exit 1
fi

#discard used parameters
shift 2

#############################################
# Wake up IPwebcam
/usr/bin/curl --max-time 3 -s http://${SRCHOST}/ -o /tmp/ipwebcam_index.html
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo -e "`date +"E%m%d %H:%M:%S.000000"`\t00000\t$0 Wake up failed" | tee -a ${LOGFILE}
	exit 1
fi
sleep 5
#############################################
# Take a new snapshot
#
/usr/bin/curl --max-time 3 -s http://${SRCHOST}/photo.jpg -o ${TMPFILE} | tee -a ${LOGFILE}
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo -e "`date +"E%m%d %H:%M:%S.000000"`\t00000\t$0 Take a new snapshot failed" | tee -a ${LOGFILE}
	exit 1
fi

#############################################
# Check tmp snapshot exist
#
if [ ! -f ${TMPFILE} ]
then
	echo -e "`date +"E%m%d %H:%M:%S.000000"`\t00000\t$0 Tmp snapshot not found ${TMPFILE}" | tee -a ${LOGFILE}
	exit 1
fi

#############################################
# Convert snapshot
#
#
# no convert : just move
#/usr/bin/convert ${SRCFILE} $@ ${DESTFILE}
mv ${TMPFILE} ${DESTFILE}
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo -e "`date +"E%m%d %H:%M:%S.000000"`\t00000\t$0 Deliver final file fail" | tee -a ${LOGFILE}
	exit 1
fi

#############################################
# Check dest file
#
if [ ! -f ${DESTFILE} ]
then
	echo -e "`date +"E%m%d %H:%M:%S.000000"`\t00000\t$0 Destination file not found ${DESTFILE}" | tee -a ${LOGFILE}
fi

#############################################
# Finish
#
echo -e "`date +"I%m%d %H:%M:%S.000000"`\t00000\t$0 Done ${DESTFILE}" | tee -a ${LOGFILE}

exit 0

