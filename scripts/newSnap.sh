#!/bin/bash
#
#set -x
#

BASEDIR=/var/goHome
SRCDIR=${BASEDIR}/motion/capture
SRCFILE=${SRCDIR}/lastsnap.jpg
DESTDIR=${BASEDIR}/www/capture
LOGFILE=${BASEDIR}/log/snapshot.log

#############################################
# Check cmde line parameter
# expecting : <motion host[:port]> <motion chanel> <dest file name> <convert parameters>
#

SRCHOST=$1
CHANEL=$2
DESTFILE=${DESTDIR}/$3

if [ "${DESTFILE}" == "${DESTDIR}/" ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Missing parameters "$@" >> ${LOGFILE}
	echo "" >> ${LOGFILE}
	echo "`basename $0` missing parameters"
	exit 1
fi

echo `date +"%Y%m%d %H%M%S $0 "` New snapshot for "$@" >> ${LOGFILE}

#discard used parameters
shift 3


#############################################
# Ask motion to take a new snapshot
#
/usr/bin/curl -s http://${SRCHOST}/${CHANEL}/action/snapshot >> ${LOGFILE}
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Call to convert fail >> ${LOGFILE}
	echo "" >> ${LOGFILE}
	echo "`basename $0` call to motion fail"
	exit 1
fi

# give time for motion to save snapshot
sleep 2

#############################################
# Check source snapshot exist (must have been created by motion)
#
if [ ! -f ${SRCFILE} ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Source snapshot not found ${SRCFILE} >> ${LOGFILE}
	echo "" >> ${LOGFILE}
	echo "`basename $0` missing snapshot"
	exit 1
fi

#############################################
# Convert snapshot
#

/usr/bin/convert ${SRCFILE} $@ ${DESTFILE}
RESULT=$?
if [ ${RESULT} != 0 ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Call to convert fail >> ${LOGFILE}
	echo "" >> ${LOGFILE}
	echo "`basename $0` convert fail"
	exit 1
fi

#############################################
# Check dest file
#
if [ ! -f ${DSTFILE} ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Destination file not found ${DSTFILE} >> ${LOGFILE}
	echo "" >> ${LOGFILE}
	echo "`basename $0` fail ${DESTFILE}"
	exit 1
fi

#############################################
# Finish
#
echo "" >> ${LOGFILE}

echo "`basename $0` Done `basename ${DESTFILE}`"

exit 0

