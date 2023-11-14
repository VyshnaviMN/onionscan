#!/bin/bash
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <file>"
    exit 1
fi

file="$1"

# Check if the file exists and is readable
if [ ! -r "$file" ]; then
    echo "File '$file' not found or not readable"
    exit 1
fi

# Loop through each line in the file and execute onionscan
while IFS= read -r line || [ -n "$line" ]; do
    echo "Scanning $line"
    ./onionscan-src -verbose -depth 1 "$line"
    echo "\n\n"
done < "$file"
