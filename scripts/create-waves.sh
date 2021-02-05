#!/bin/bash

FILES="am.csv  decay.csv  fsk.csv  full.csv  lorenz.csv  mtone.csv  square.csv  triangle.csv noise.csv"
for F in $FILES
do
    ./bin/mhs5200a convert "/data/src/electronics/mhs5200a/waves/${F}" >"waves/${F}"
done
