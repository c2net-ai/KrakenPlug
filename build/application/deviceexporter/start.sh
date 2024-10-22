#!/bin/bash

source /app/copy_lib.sh

set -- /app/deviceexporter "$@"
exec "$@"
