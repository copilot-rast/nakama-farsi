#!/bin/bash
P=${CL_PREFIX:-nkce2e}; RP=${REDIS_PORT:-6495}
R="/opt/homebrew/bin/redis-cli -p $RP"
L="local n=0 local c=0 local cur='0' repeat local r=redis.call('SCAN',cur,'MATCH',ARGV[1],'COUNT',1000) cur=r[1] for _,k in ipairs(r[2]) do n=n+1 c=c+redis.call('HLEN',k) end until cur=='0' return {n,c}"
echo "presence_keys_and_entries=$($R eval "$L" 0 "$P:{*}:p*" | tr '\n' ' ')"
echo "ticket_keys_and_entries=$($R eval "$L" 0 "$P:mm:*:t" | tr '\n' ' ')"
echo "redis_mb=$($R info memory | awk -F: '/^used_memory:/{printf "%.1f",$2/1048576}')"
