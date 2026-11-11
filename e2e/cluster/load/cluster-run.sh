#!/bin/bash
set -e
HERE=$(cd "$(dirname "$0")" && pwd)
V=$1; RAMP=$2; HOLD=$3; L=$4; NODES=${NODES:-"1 2 3 4"}
export RUN=${RUN:-$HERE/../.run} CL_PREFIX=${CL_PREFIX:-nkce2e} MM_PROP=${MM_PROP:-mode} MM_SHARDS=${MM_SHARDS:-8} LEVEL=${LEVEL:-warn}
export MODS=${MODS:-$RUN/modules}
OUT=$RUN/load; mkdir -p $OUT
T=""
for i in $NODES; do for ip in 127.0.0.1 10.255.254.34 100.106.1.79; do T="$T,$ip:$((7600+i*10+1))"; done; done
export TARGETS=${T#,}
export PATH=$HOME/.nvm/versions/node/v22.18.0/bin:$PATH
(sleep $((RAMP+15)); cd $HERE/.. && DUR=$((HOLD-30)) node load/probe.mjs > $OUT/probe$L.json 2>&1) &
(sleep $((RAMP+HOLD/2)); cd $HERE/.. && PREFIX=ld$L VUS=$V node load/online.mjs > $OUT/online$L.json 2>&1) &
NODES="$NODES" ./run.sh $V ld$L $RAMP $HOLD $L
sleep 20
(cd $HERE/.. && PREFIX=ld$L VUS=$V node load/online.mjs > $OUT/after$L.json 2>&1)
./ghosts.sh > $OUT/ghosts$L.txt
for i in $NODES; do grep -c '"level":"error"' $RUN/n$i/out.log; done > $OUT/errors$L.txt || true
wait
