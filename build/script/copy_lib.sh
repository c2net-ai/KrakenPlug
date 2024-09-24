#!/bin/bash

libs=(
  libcndev.so
  libefml.so
  libnvidia-ml.so
)

function copy_lib() {
  for target in $(find "/host/usr/" -name "${1}" | grep -v "stubs"); do
#    if [[ $(objdump -p ${target} 2>/dev/null | grep -o "SONAME") == "SONAME" ]]; then
      ls -l ${target}
      cp ${target} "/usr/lib64"
#    fi
  done
}

for lib in ${libs[@]}; do
  copy_lib ${lib}
done