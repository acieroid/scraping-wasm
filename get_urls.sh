#!/bin/sh
URL_SOURCE="https://downloads.majestic.com/majestic_million.csv"
echo 'Downloading URLs...'
curl "$URL_SOURCE" -o urls.csv
cut -d, -f 3 urls.csv | sed 1d | sed 's|^|http://|' > urls.txt

N=$(wc -l urls.txt | cut -d' ' -f1)
echo "There are $N top-level URLs to scrape"
