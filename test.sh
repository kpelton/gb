#!/bin/bash
runtime=${1:-5s}
lines="200000"
rom="roms/Tetris.gb"
compare="diffuse"
./main $rom > main_out &
./gbem $rom > gbem_out &
#Sleep for the specified time.
sleep $runtime
killall main
killall gbem

tail -n+7 gbem_out | head -n $lines > gbem_out2
head -n $lines main_out  > main_out2
mv gbem_out2 gbem_out
mv main_out2 main_out

$compare main_out gbem_out