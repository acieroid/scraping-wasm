#!/bin/sh
# Script that counts how many modules export their memory

DIR=bytecode
WASSAIL=wassail

MEM_EXPORTS=0
NO_MEM_EXPORTS=0
for module in $(ls $DIR/*.wasm)
do
    echo "Processing $module"
    HASH=$(basename $module | cut -d'.' -f1)
    EXPORTCOUNT=$($WASSAIL mem-exports $module)
    if [ "$EXPORTCOUNT" -ge 1 ]; then
        MEM_EXPORTS=$(echo "$MEM_EXPORTS+1" | bc)
    else
        NO_MEM_EXPORTS=$(echo "$NO_MEM_EXPORTS+1" | bc)
    fi
done
echo "$MEM_EXPORTS modules export their memory, while $NO_MEM_EXPORTS do not"
