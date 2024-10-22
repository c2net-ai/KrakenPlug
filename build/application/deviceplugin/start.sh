#!/bin/bash

source /app/copy_lib.sh

set -- /app/deviceplugin "$@"
exec "$@"
