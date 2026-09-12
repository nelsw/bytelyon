#!/bin/bash

echo "Enter a ꟛƒ name [news|page|serp]:"
read -r key
aws logs tail "/aws/lambda/bytelyon-$key-scraper" --follow
