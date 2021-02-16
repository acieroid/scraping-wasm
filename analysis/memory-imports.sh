#!/bin/sh
# Script that counts how many modules import their memory

DIR=bytecode
WASSAIL=wassail

MEM_IMPORTS=0
NO_MEM_IMPORTS=0
for module in $(ls $DIR/*.wasm)
do
    echo "Processing $module"
    HASH=$(basename $module | cut -d'.' -f1)
    IMPORTCOUNT=$($WASSAIL mem-imports $module)
    if [ "$IMPORTCOUNT" -ge 1 ]; then
        MEM_IMPORTS=$(echo "$MEM_IMPORTS+1" | bc)
    else
        NO_MEM_IMPORTS=$(echo "$NO_MEM_IMPORTS+1" | bc)
    fi
done
echo "$MEM_IMPORTS modules import their memory, while $NO_MEM_IMPORTS do not"
