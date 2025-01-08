#!/bin/bash

#source /app/copy_lib.sh > /dev/null 2>&1
ldconfig
set -- /app/deviceplugin "$@"
exec "$@"
