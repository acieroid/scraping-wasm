#!/bin/sh
cut -d, -f1 results.csv | uniq -c | sort -g
