#!/bin/sh
# This scripts combines the results from multiple scrapings (as zip of .log files) into single .log files

# Extract all zip files
for archive in $(ls *.zip)
do
    unzip -o $archive
done

# For each type of file
for file in $(echo scripts.log noscripts.log dnserrors.log failures.log timeouts.log)
do
    # Find all corresponding results file
    for result_file in $(find . -name $file)
    do
        if [ "$result_file" != "./$file" ]
        then
            # And append them to the file that contains all results
            # The sed commands perform the following:
            #   Add http: in front of lines starting with //
            #   Remove lines starting with / only (these are invalid links)
            #   Only keep lines starting with http
            cat $result_file | sed 's|^//|http://|' | sed '/^\/[^/]/d' | sed -n '/http:/p' >> $file
        fi
    done
done

# Uniq all files
export LC_ALL=C
for file in $(echo scripts.log noscripts.log dnserrors.log failures.log timeouts.log)
do
    sort $file | uniq > $file.2
    mv $file.2 $file
done
