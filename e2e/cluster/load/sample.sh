#!/bin/bash
OUT=$1; RUN=${RUN:?run dir}; NODES=${NODES:-"1"}; INT=${INT:-5}; RPORT=${RPORT:-6495}
echo "ts,node,rss_mb,cpu,goroutines,heap_mb,sessions,presences,redis_mb,redis_ops" > $OUT
while [ -f $RUN/sampling ]; do
  ts=$(date +%s)
  rinfo=$(/opt/homebrew/bin/redis-cli -p $RPORT info 2>/dev/null | tr -d '\r')
  rm=$(echo "$rinfo" | awk -F: '/^used_memory:/{printf "%.1f",$2/1048576}')
  ro=$(echo "$rinfo" | awk -F: '/^instantaneous_ops_per_sec:/{print $2}')
  for i in $NODES; do
    P=$(cat $RUN/n$i/pid); kill -0 $P 2>/dev/null || continue
    read rss cpu <<<"$(ps -o rss=,%cpu= -p $P)"
    M=$(curl -s -m 2 localhost:$((7600+i*10+4))/metrics)
    g=$(echo "$M" | awk '/^go_goroutines /{print $2}')
    h=$(echo "$M" | awk '/^go_memstats_heap_inuse_bytes /{printf "%.1f",$2/1048576}')
    se=$(echo "$M" | awk '/^nakama_sessions[{ ]/{print $2}')
    pr=$(echo "$M" | awk '/^nakama_presences[{ ]/{print $2}')
    echo "$ts,node$i,$((rss/1024)),$cpu,$g,$h,$se,$pr,${rm:-0},${ro:-0}" >> $OUT
  done
  sleep $INT
done
