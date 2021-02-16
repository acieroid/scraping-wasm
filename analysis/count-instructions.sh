#!/bin/sh
# Script that counts instructions used by WebAssembly modules
DIR=bytecode
WASSAIL=wassail

for module in $(ls $DIR/*.wasm)
do
    echo "Processing $module"
    $WASSAIL instructions $module | cut -f1,3 >> counts.tmp
done

awk ' { tot[$2] += $1 } END { for (i in tot) print tot[i],i } ' counts.tmp | sort -g

rm counts.tmp
