#!/bin/sh
# For each wasm module, gives the count of how many website use it

DIR=./bytecode/
RESULTSCSV=results.csv
for module in $(ls $DIR/*.wasm)
do
    HASH=$(basename $module | cut -d'.' -f1)
    SITES_THAT_USE_IT=$(grep "$HASH" "$RESULTSCSV" | cut -d',' -f1 | sort | uniq | wc -l)
    echo -e "$SITES_THAT_USE_IT\t$HASH"
done
