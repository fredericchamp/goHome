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

#############################################
# Check cmde line parameter
# expecting : <host:port> <dest file name> <convert parameters>
#

SRCHOST=$1
DESTFILE=${DESTDIR}/$2

if [ "${DESTFILE}" == "${DESTDIR}/" ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Missing parameters "$@" | tee -a ${LOGFILE}
	exit 1
fi

echo `date +"%Y%m%d %H%M%S $0 "` New IPwebcam snap shot for "$@" | tee -a ${LOGFILE}

#discard used parameters
shift 2

#############################################
# Weak up IPwebcam
/usr/bin/curl -s http://${SRCHOST}/ -o /tmp/ipwebcam_index.html | tee -a ${LOGFILE}
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Weak up failed | tee -a ${LOGFILE}
	exit 1
fi
sleep 5
#############################################
# Take a new snapshot
#
/usr/bin/curl -s http://${SRCHOST}/photo.jpg -o ${TMPFILE} | tee -a ${LOGFILE}
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Take a new snapshot failed | tee -a ${LOGFILE}
	exit 1
fi

#############################################
# Check tmp snapshot exist
#
if [ ! -f ${TMPFILE} ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Tmp snapshot not found ${TMPFILE} | tee -a ${LOGFILE}
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
	echo `date +"%Y%m%d %H%M%S $0 "` Deliver final file fail | tee -a ${LOGFILE}
	exit 1
fi

#############################################
# Check dest file
#
if [ ! -f ${DSTFILE} ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Destination file not found ${DSTFILE} | tee -a ${LOGFILE}
	exit 1
fi

#############################################
# Finish
#
echo "`basename $0` Done `basename ${DESTFILE}`" | tee -a ${LOGFILE}

exit 0

