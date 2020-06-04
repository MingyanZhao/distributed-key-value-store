#!/bin/bash

function run_test {
  # TODO: use the docker container to bring up the env
  test_spec=${1}

  bin/leader --caster=simple &
  LEADER_PID=$!
  bin/follower --follower_id=0 &
  FOLLOWER0_PID=$!
  bin/follower --follower_id=1 &
  FOLLOWER1_PID=$!

  bin/test_client --test_spec=${test_spec}

  if [ $? -eq 0 ]
  then
    echo -e "\e[1m \e[32m Test ${test_spec} Succeeded\e[0m"
  else
    echo -e "\e[1m \e[31m Test ${test_spec} Failed\e[0m"
  fi
  kill $LEADER_PID
  kill $FOLLOWER0_PID
  kill $FOLLOWER1_PID
}

if [[ ! -f "bin/leader" ]] && [[ ! -f "bin/follower" ]] && [[ ! -f "bin/test_client" ]]; then
  echo "Run 'make' first"
  exit
fi

for test_spec in test/*.pb.txt; do
  run_test $test_spec
done
