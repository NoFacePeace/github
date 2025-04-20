#!/bin/bash
echo -e "PID\tUSER\tCMD\t%CPU\t%MEM\tOPEN_FILES"
for pid in $(ps -e -o pid=  --sort=-%cpu | head -n 20); do
  if [ -d /proc/$pid ]; then
    user=$(stat -c '%U' /proc/$pid)
    cmd=$(tr -d '\0' < /proc/$pid/cmdline)
    cpu_mem=$(ps -p $pid -o %cpu,%mem --no-headers)
    open_files=$(ls /proc/$pid/fd 2>/dev/null | wc -l)
    echo -e "$pid\t$user\t${cmd:0:128}\t$cpu_mem\t$open_files"
  fi
done
