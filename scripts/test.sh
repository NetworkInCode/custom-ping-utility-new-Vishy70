#!/bin/bash

# This script is a placeholder for testing the functionality of the custom ping utility.
# Tests will be added later.
cd $1
mkdir ../bin

go build -o ../bin/pinger
cd ../bin

sudo ./pinger -I wlp45s0 -c 4 -4 nitk.ac.in