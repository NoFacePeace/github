#!/bin/bash
dir="/data/log/monitor/"
file=$(date "+%Y%m%d")
now=$(date "+%Y-%m-%d %H:%M:%S")


# 网络
echo ${now} >> "${dir}netstat_${file}.log"
netstat -tunp >> "${dir}netstat_${file}.log"


# 进程 top
echo ${now} >> "${dir}top_${file}.log"
top -b | head -n 50 >> "${dir}top_${file}.log"

# fd
echo ${now} >> "${dir}ps_${file}.log"
echo -e "PID\tUSER\tCMD\t%CPU\t%MEM\tOPEN_FILES" >> "${dir}ps_${file}.log"
for pid in $(ps -e -o pid=  --sort=-%cpu | head -n 50); do
  if [ -d /proc/$pid ]; then
    user=$(stat -c '%U' /proc/$pid)
    cmd=$(tr -d '\0' < /proc/$pid/cmdline)
    cpu_mem=$(ps -p $pid -o %cpu,%mem --no-headers)
    open_files=$(ls /proc/$pid/fd 2>/dev/null | wc -l)
    echo -e "$pid\t$user\t${cmd:0:128}\t$cpu_mem\t$open_files" >> "${dir}ps_${file}.log"
  fi
done

# 进程 IO top
echo ${now} >> "${dir}iotop_${file}.log"
iotop -b -o -n 1 >> "${dir}iotop_${file}.log"

# io
echo ${now} >> "${dir}iostat_${file}.log"
iostat -xmd >> "${dir}iostat_${file}.log"