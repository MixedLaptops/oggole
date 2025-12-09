#!/bin/bash

# Search performance testing script

echo "=== Search Performance Test ==="
echo ""

SEARCH_TERMS=("software" "development" "DevOps" "continuous" "encyclopedia")

echo "Testing with current implementation (ILIKE)..."
echo ""

for term in "${SEARCH_TERMS[@]}"; do
    echo -n "Searching for '$term': "
    
    # Measure time using time command, extract milliseconds
    START=$(date +%s%3N)
    curl -s "http://localhost:8080/api/search?q=$term&language=en" > /dev/null
    END=$(date +%s%3N)
    
    DURATION=$((END - START))
    echo "${DURATION}ms"
done

echo ""
echo "Average response time across all queries"
