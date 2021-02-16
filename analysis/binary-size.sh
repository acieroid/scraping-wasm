#!/bin/sh
# Script that extracts the statistical distribution of the sizes of a set of files

DIR=bytecode

ls -l $DIR | tail -n +2 | awk '{print $5}' > sizes.txt

MIN=$(awk 'NR==1||$0<x{x=$0}END{print x}' sizes.txt)
MAX=$(awk 'NR==1||$0>x{x=$0}END{print x}' sizes.txt)
MEDIAN=$(sort -n sizes.txt | awk '{a[NR]=$0}END{print(NR%2==1)?a[int(NR/2)+1]:(a[NR/2]+a[NR/2+1])/2}')
MEAN=$(awk '{x+=$0}END{print x/NR}' sizes.txt)

echo "min: $MIN bytes"
echo "max: $MAX bytes"
echo "median: $MEDIAN bytes"
echo "mean: $MEAN bytes"
