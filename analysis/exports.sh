#!/bin/sh
# Script that links each exported functions to the files that export it

DIR=bytecode
WASSAIL=wassail

# Produce an export-modules.txt file, in which each line is an export by a given file, in the following format:
# file    exportname     exporttype
echo -ne > export-modules.txt
for module in $(ls $DIR/*.wasm)
do
    echo "Procesing $module"
    HASH=$(basename $module | cut -d'.' -f1)
    $WASSAIL exports $module | cut -f2-3 | sed "s/^/$HASH\t/" >> exports-modules.txt
done

# Produces an export.stats file, in which each line is an export and by how many modules it is exported, in the following format:
# n-usages    exportname     exportype
echo > export-usages.txt
cat export-modules.txt | cut -f2-3 | sort | uniq | while read export
do
    COUNT=$(grep "$export" export-modules.txt | wc -l)
    echo -ne "$COUNT $export\n" >> export-usages.txt
done

echo "Produced 2 file:"
echo -e "\texport-modules.txt lists exports of each module, one export per line"
echo -e "\texport-usages counts total number of usage per export, one export per line"
