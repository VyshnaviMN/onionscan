file="$1"

while IFS= read -r line || [ -n "$line" ]; do
    echo "\nScanning $line"
done < "$file"
echo "Done!"