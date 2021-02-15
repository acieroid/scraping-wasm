#!/bin/sh
# Script that extracts the sizes of WebAssembly module sections

DIR=bytecode
WASSAIL=wassail

echo -ne > sizes.txt
for module in $(ls $DIR/*.wasm)
do
    echo "Processing $module"
    $WASSAIL sizes $module >> sizes.txt
done

for typ in type import func table memory global export start elem code data
do
    MIN=$(grep $typ sizes.txt | cut -f1 | awk 'NR==1||$0<x{x=$0}END{print x}')
    MAX=$(grep $typ sizes.txt | cut -f1 | awk 'NR==1||$0>x{x=$0}END{print x}')
    MEDIAN=$(grep $typ sizes.txt | cut -f1 | sort -n | awk '{a[NR]=$0}END{print(NR%2==1)?a[int(NR/2)+1]:(a[NR/2]+a[NR/2+1])/2}')
    MEAN=$(grep $typ sizes.txt | cut -f1 | awk '{x+=$0}END{print x/NR}')
    echo -e "$typ:"
    echo -e "\tmin: $MIN"
    echo -e "\tmax: $MAX"
    echo -e "\tmedian: $MEDIAN"
    echo -e "\tmean: $MEAN"
done
#rm sizes.txt
