#!/bin/sh
# Script that links each exported functions to the files that export it

DIR=bytecode
WASSAIL=wassail

# Produce an exports-modules.txt file, in which each line is an export by a given file, in the following format:
# file    exportname     exporttype
echo -ne > export-modules.txt
echo -ne > exports-per-module.txt
for module in $(ls $DIR/*.wasm)
do
    echo "Procesing $module"
    HASH=$(basename $module | cut -d'.' -f1)
    ALLEXPORTS=$($WASSAIL exports $module)
    echo -n "$ALLEXPORTS" | cut -f2-3 | sed "s/^/$HASH\t/" >> export-modules.txt
    echo "$ALLEXPORTS" | wc -l | sed "s/$/\t$HASH/" >> exports-per-module.txt
done

# Produces an export.stats file, in which each line is an export and by how many modules it is used, in the following format:
# n-usages    exportname     exporttype
echo -ne > export-usages.txt
cat export-modules.txt | cut -f2-3 | sort | uniq | while read export
do
    COUNT=$(grep -F -- "$export" export-modules.txt | wc -l)
    echo -ne "$COUNT $export\n" >> export-usages.txt
done

echo "Produced the following files:"
echo -e "\texport-modules.txt lists exports of each module, one export per line"
echo -e "\texport-usages.txt counts total number of usage per export, one export per line"
echo -e "\texports-per-module.txt counts the number of exports per module, one module per line"
