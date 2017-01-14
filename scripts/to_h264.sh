#!/bin/bash
#
#set -x
#

BASEDIR=/var/goHome
LOGFILE=${BASEDIR}/log/transcode.log
DESTDIR=${BASEDIR}/motion/capture

#############################################
# Check cmde line parameter
#
if [ -z $1 ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Missing source file name >> ${LOGFILE}
	echo "" >> ${LOGFILE}
	exit 1
fi

SRCFILE=$1
SRCFILENAME=`basename $1`
DSTFILENAME=${SRCFILENAME%.avi}.mp4
DSTFILE=${DESTDIR}/${DSTFILENAME}

echo `date +"%Y%m%d %H%M%S $0 "` Transcode ${SRCFILE} to ${DSTFILE} >> ${LOGFILE}

#############################################
# Check source file exist
#
if [ ! -f ${SRCFILE} ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Source file not found ${SRCFILE} >> ${LOGFILE}
	echo "" >> ${LOGFILE}
	exit 1
fi

#############################################
# Tanscode video file
#

DSTART=`date +%s`

avconv -i ${SRCFILE} -y -an -c:v:0 libx264 -c:a copy ${DSTFILE}

DEND=`date +%s`

#############################################
# Check dest file and remove temp copy
#
if [ ! -f ${DSTFILE} ]
then
	echo `date +"%Y%m%d %H%M%S $0 "` Destination file not found ${DSTFILE} >> ${LOGFILE}
	echo "" >> ${LOGFILE}
	exit 1
fi

#############################################
# Cleanup : remove source file
#

rm -f ${SRCFILE}

#############################################
# Finish
#
echo `date +"%Y%m%d %H%M%S $0 "` Transcode done in $((${DEND} - ${DSTART})) sec >> ${LOGFILE}
echo "" >> ${LOGFILE}


exit 0

