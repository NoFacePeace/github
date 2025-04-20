#!/bin/bash
file=$(date "+%Y%m%d")
now=$(date "+%Y-%m-%d %H:%M:%S")

echo ${now} >> "netstat_${file}.log"
netstat -tunp >> "netstat_${file}.log"


