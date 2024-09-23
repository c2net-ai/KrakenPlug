#!/bin/bash

/app/copy_lib.sh

set -- /app/deviceplugin "$@"
exec "$@"
