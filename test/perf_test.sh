#!/bin/bash

THREADCOUNT=100
REQUESTSCOUNT=600
TESTTIME=4

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
  # Perf test parameters

  bin/client --follower_id=0 --mode=perf --config_file=${config_file} --threadcount=$THREADCOUNT --requestcount=$REQUESTSCOUNT --testtime=$TESTTIME &
  CLIENT0_PID=$!
  bin/client --follower_id=1 --mode=perf --config_file=${config_file} --threadcount=$THREADCOUNT --requestcount=$REQUESTSCOUNT --testtime=$TESTTIME &
  CLIENT1_PID=$!
  bin/client --follower_id=2 --mode=perf --config_file=${config_file} --threadcount=$THREADCOUNT --requestcount=$REQUESTSCOUNT --testtime=$TESTTIME &
  CLIENT2_PID=$!
  bin/client --follower_id=3 --mode=perf --config_file=${config_file} --threadcount=$THREADCOUNT --requestcount=$REQUESTSCOUNT --testtime=$TESTTIME &
  CLIENT3_PID=$!
  bin/client --follower_id=4 --mode=perf --config_file=${config_file} --threadcount=$THREADCOUNT --requestcount=$REQUESTSCOUNT --testtime=$TESTTIME &
  CLIENT4_PID=$!
  bin/client --follower_id=5 --mode=perf --config_file=${config_file} --threadcount=$THREADCOUNT --requestcount=$REQUESTSCOUNT --testtime=$TESTTIME &
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

# Sleep for test duration
sleep $TESTTIME

# Sleep for 10 more  seconds to wait for the client to finish
sleep 10

echo "Kill Servers"
kill $FOLLOWER0_PID
kill $FOLLOWER1_PID
kill $FOLLOWER2_PID
kill $FOLLOWER3_PID
kill $FOLLOWER4_PID
kill $FOLLOWER5_PID
kill $LEADER_PID

echo "Comparing results..."
echo "diff 0 to 1..."
diff perfresult-follower-0 perfresult-follower-1 > diff-0-1.txt
echo "diff 0 to 2..."
diff perfresult-follower-0 perfresult-follower-2 > diff-0-2.txt
echo "diff 0 to 3..."
diff perfresult-follower-0 perfresult-follower-3 > diff-0-3.txt
echo "diff 0 to 4..."
diff perfresult-follower-0 perfresult-follower-4 > diff-0-4.txt
echo "diff 0 to 5..."
diff perfresult-follower-0 perfresult-follower-5 > diff-0-5.txt

echo "All done!"