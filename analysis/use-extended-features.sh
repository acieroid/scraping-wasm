#!/bin/sh
# Count how many modules use features that are not part of the default standard
# It is possible to refine the results using the find-extensions.sh script
DIR=bytecode
WASM2WAT=wasm2wat
ARGS_NOEXT="--disable-mutable-globals --disable-saturating-float-to-int --disable-sign-extension --disable-multi-value"
VALID=0
INVALID=0

for module in $(ls $DIR/*.wasm)
do
    echo "Processing $module"
    $WASM2WAT $ARGS_NOEXT $module >/dev/null
    if [ "$?" -eq "0" ]
    then
        VALID=$(echo $VALID+1 | bc)
    else
        INVALID=$(echo $INVALID+1 | bc)
        echo "module $module is invalid"
    fi
done
echo "$VALID modules do not use extended features, while $INVALID do"
