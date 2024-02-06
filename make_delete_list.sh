#!/bin/bash 

cat senders.json | jq -r 'keys[] as $k | "\($k) \(.[$k] | length)"' > delete_list.txt
