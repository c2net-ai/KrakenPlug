#!/bin/bash

target_dir="/usr/lib"
libs=("libcndev.so" "libefml.so" "libnvidia-ml.so")
search_dirs=("/host/usr/lib/x86_64-linux-gnu" "/host/usr/local/neuware/lib64" "/host/usr/local/Ascend/driver/lib64" "/host/usr/lib" "/host/usr/lib64" )

function search_and_copy {
    local lib=$1
    for dir in "${search_dirs[@]}"; do
        for file_path in $(find "$dir" -name "$lib"); do
            echo "Found $lib in $file_path"
            cp "$file_path" "$target_dir"
            return 0
        done
    done
    return 1
}

for lib in "${libs[@]}"; do
    search_and_copy "$lib"
#    if ! search_and_copy "$lib"; then
#        break
#    fi
done

