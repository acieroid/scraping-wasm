#!/bin/sh
# Script that links each imported functions to the files that import it

DIR=bytecode
WASSAIL=wassail

# Produce an imports-modules.txt file, in which each line is an import by a given file, in the following format:
# file    importname     importtype
echo -ne > import-modules.txt
for module in $(ls $DIR/*.wasm)
do
    echo "Procesing $module"
    HASH=$(basename $module | cut -d'.' -f1)
    $WASSAIL imports $module | cut -f2-3 | sed "s/^/$HASH\t/" >> import-modules.txt
done

# Produces an import.stats file, in which each line is an import and by how many modules it is used, in the following format:
# n-usages    importname     importtype
echo > import-usages.txt
cat import-modules.txt | cut -f2-3 | sort | uniq | while read import
do
    COUNT=$(grep "$import" import-modules.txt | wc -l)
    echo -ne "$COUNT $import\n" >> import-usages.txt
done

echo "Produced 2 file:"
echo -e "\timport-modules.txt lists imports of each module, one import per line"
echo -e "\timport-usages counts total number of usage per import, one import per line"
