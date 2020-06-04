#!/bin/bash

function startFollowers {
  # TODO: use the docker container to bring up the env
  config_file=${1}

  bin/leader --caster=simple --config_file=${config_file} &
  LEADER_PID=$!

  bin/follower --follower_id=0 --config_file=${config_file} &
  FOLLOWER0_PID=$!
  bin/follower --follower_id=1 --config_file=${config_file} &
  FOLLOWER1_PID=$!
  bin/follower --follower_id=2 --config_file=${config_file} &
  FOLLOWER2_PID=$!
  bin/follower --follower_id=3 --config_file=${config_file} &
  FOLLOWER3_PID=$!
  bin/follower --follower_id=4 --config_file=${config_file} &
  FOLLOWER4_PID=$!
  bin/follower --follower_id=5 --config_file=${config_file} &
  FOLLOWER5_PID=$!

  echo "Servers PIDs"
  echo $LEADER_PID
  echo $FOLLOWER0_PID
  echo $FOLLOWER1_PID
  echo $FOLLOWER2_PID
  echo $FOLLOWER3_PID
  echo $FOLLOWER4_PID
  echo $FOLLOWER5_PID
}

function startClients {
  config_file=${1}

  bin/client --follower_id=0 --mode=perf --config_file=${config_file}  &
  CLIENT0_PID=$!
  bin/client --follower_id=1 --mode=perf --config_file=${config_file} > c-1.log &
  CLIENT1_PID=$!
  bin/client --follower_id=2 --mode=perf --config_file=${config_file} > c-2.log &
  CLIENT2_PID=$!
  bin/client --follower_id=3 --mode=perf --config_file=${config_file} > c-3.log &
  CLIENT3_PID=$!
  bin/client --follower_id=4 --mode=perf --config_file=${config_file} > c-4.log &
  CLIENT4_PID=$!
  bin/client --follower_id=5 --mode=perf --config_file=${config_file} > c-5.log &
  CLIENT5_PID=$!

  echo "Clients PIDs"
  echo $CLIENT0_PID
  echo $CLIENT1_PID
  echo $CLIENT2_PID
  echo $CLIENT3_PID
  echo $CLIENT4_PID
  echo $CLIENT5_PID
}

if [[ ! -f "bin/leader" ]] && [[ ! -f "bin/follower" ]] && [[ ! -f "bin/test_client" ]]; then
  echo "Run 'make' first"
  exit
fi
pwd

CONFIG=config/perf-local-config.pb.txt

startFollowers $CONFIG

sleep 3
echo "Followers started"

startClients $CONFIG

# Sleep for 15 seconds to wait for the client to finish.log 
# Clients are waiting for 10 seconds after sending the requests to wait for leader to broadcasting.
sleep 15

echo "Kill Servers"
kill $LEADER_PID
kill $FOLLOWER0_PID
kill $FOLLOWER1_PID
kill $FOLLOWER2_PID
kill $FOLLOWER3_PID
kill $FOLLOWER4_PID
kill $FOLLOWER5_PID

echo "Comparing results..."
diff perfresult-follower-0 perfresult-follower-1
diff perfresult-follower-0 perfresult-follower-2
diff perfresult-follower-0 perfresult-follower-3
diff perfresult-follower-0 perfresult-follower-4
diff perfresult-follower-0 perfresult-follower-5

echo "All done!"