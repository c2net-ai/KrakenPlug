#!/bin/bash

ARCHIVE=`awk '/^__ARCHIVE_BOUNDARY__/ { print NR + 1; exit 0; }' $0`

tail -n +$ARCHIVE $0 > kptools.tar.gz
tar -zpxf kptools.tar.gz
cp kptools/kpsmi /usr/local/bin
cp kptools/kprunc /usr/local/bin
mkdir -p /etc/kprunc
cp kptools/config.yaml /etc/kprunc
rm kptools.tar.gz
rm -rf kptools

exit 0
__ARCHIVE_BOUNDARY__ #脚本最后加一空行
