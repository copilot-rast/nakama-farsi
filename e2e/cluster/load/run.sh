#!/bin/bash
cd "$(dirname "$0")"
V=$1; PF=$2; RAMP=$3; HOLD=$4; L=$5; OUT=${OUT:-$RUN/load}
mkdir -p $OUT; touch $RUN/sampling
./sample.sh $OUT/l$L.csv &
k6 run -q --no-color -e TARGETS="$TARGETS" -e VUS=$V -e RAMP=${RAMP}s -e HOLD=$HOLD -e PREFIX=$PF -e SKEY="${SKEY:-defaultkey}" --summary-trend-stats "avg,p(50),p(95),max" load.js > $OUT/k6-$L.txt 2>&1
rm -f $RUN/sampling; wait; python3 summ.py $OUT/l$L.csv
