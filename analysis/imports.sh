#!/bin/sh
# Script that links each imported functions to the files that import it

DIR=bytecode
WASSAIL=wassail

# Produce an imports-modules.txt file, in which each line is an import by a given file, in the following format:
# file    importname     importtype
echo -ne > import-modules.txt
echo -ne > imports-per-module.txt
for module in $(ls $DIR/*.wasm)
do
    echo "Procesing $module"
    HASH=$(basename $module | cut -d'.' -f1)
    ALLIMPORTS=$($WASSAIL imports $module)
    echo -n "$ALLIMPORTS" | cut -f2-3 | sed "s/^/$HASH\t/" >> import-modules.txt
    echo "$ALLIMPORTS" | wc -l | sed "s/$/\t$HASH/" >> imports-per-module.txt
done

# Produces an import.stats file, in which each line is an import and by how many modules it is used, in the following format:
# n-usages    importname     importtype
echo -ne > import-usages.txt
cat import-modules.txt | cut -f2-3 | sort | uniq | while read import
do
    COUNT=$(grep -F -- "$import" import-modules.txt | wc -l)
    echo -ne "$COUNT $import\n" >> import-usages.txt
done

echo "Produced the following files:"
echo -e "\timport-modules.txt lists imports of each module, one import per line"
echo -e "\timport-usages.txt counts total number of usage per import, one import per line"
echo -e "\timports-per-module.txt counts the number of imports per module, one module per line"
