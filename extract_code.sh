#!/bin/bash 

echo $1 | grep -o -E 'code\\=[^&\\]*' | cut -d '=' -f 2 
