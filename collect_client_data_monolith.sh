#!/bin/sh

# This assumes that terraform has deployed everything

# setup env vars
OUTPUT=$(terraform -chdir=terraform output -json service_endpoints)
CLIENT_HOST=$(echo $OUTPUT | jq -r .client)

ENGINE_HOST=$(echo $OUTPUT | jq -r .monolith)
MUSIC_HOST=$(echo $OUTPUT | jq -r .monolith)
INITIATOR_HOST=$(echo $OUTPUT | jq -r .monolith)
WORLDGEN_HOST=$(echo $OUTPUT | jq -r .monolith)
AUTH_HOST=$(echo $OUTPUT | jq -r .monolith)
SCORE_HOST=$(echo $OUTPUT | jq -r .monolith)

SPKI_HASH=$(cat ./certs/spki_hash.txt)
DEPLOY_TIME=$(cat ./terraform/deploy_time.txt)
DEPLOY_TYPE=$(cat ./terraform/deploy_type.txt)
TIMESTAMP=collected_$(date +"%Y%m%d_%H%M%S")

# svcs

# worldgen 
# auth
# engine
# initiator
# music
# score

restart_svc() {
    ssh -i terraform/certs/ssh_key ec2-user@$1 <<EOF
    sudo systemctl restart flappygo
EOF

}

run_for_seed() {
    # https://stackoverflow.com/questions/4412238/what-is-the-cleanest-way-to-ssh-and-run-multiple-commands-in-bash
    # need to be root!
    ssh -i terraform/certs/ssh_key ec2-user@$WORLDGEN_HOST <<EOF
        sudo mkdir -p /etc/systemd/system/flappygo.service.d/
        # Hard mode
        sudo sh -c 'echo -e "[Service]\nEnvironment=\"STABLE_WORLD_SEED=${1}\"" > /etc/systemd/system/flappygo.service.d/seed.conf'
        sudo cp /opt/flappygo/backend/statout/stats.csv /opt/flappygo/backend/statout/stats_old.csv
        sudo systemctl daemon-reload
        sudo systemctl restart flappygo
EOF

}

run_iteration() {
    cd client-automation

    # you must set the GAME_HOST
    STAT_DIR=../stat/$DEPLOY_TIME/$TIMESTAMP/client_seed_${1}_run_${2} INPUT_CSV=input_seeds/${1}.csv GAME_URL=https://${CLIENT_HOST} poetry run python src/input_simulator.py --origin-to-force-quic-on=$ENGINE_HOST:4433,$MUSIC_HOST:4433 --ignore-certificate-errors-spki-list=${SPKI_HASH}

    cd ..

}

# ../stat/$DEPLOY_TIME/$TIMESTAMP/client_seed_${1}_run_${2} 

# https://stackoverflow.com/questions/49110/how-do-i-write-a-for-loop-in-bash
run_test() {
    run_for_seed $1
    for i in $(seq 1 5); do
        echo "Run $i seed $1"

        score=0
        while [ "$score" -eq 0 ]
        do
            if test -d "./stat/$DEPLOY_TIME/$TIMESTAMP/client_seed_${1}_run_${i}/"; then 
                rm -rf "./stat/$DEPLOY_TIME/$TIMESTAMP/client_seed_${1}_run_${i}/"
            fi

            run_iteration $1 $i
            score=$(cat ./stat/$DEPLOY_TIME/$TIMESTAMP/client_seed_${1}_run_${i}/extra_data.json | jq .score)
            # https://serverfault.com/questions/7503/how-to-determine-if-a-bash-variable-is-empty
            if [[ -z "${score}" ]]; then
                score=0
            fi
            echo "Got a score of $score"
        done
    done

}

# this is for resetting logs 
echo "Restarting svcs"

# 1 is good
restart_svc $WORLDGEN_HOST

sleep 5

echo "Restarted svcs"

mkdir -p ./stat/$DEPLOY_TIME

# could be done earlier, but whatever
echo $DEPLOY_TYPE > ./stat/$DEPLOY_TIME/deploy_type

run_test 8525333463046388971
run_test 6977347407732442987

# done, now collect stats over all the runs
./collect_remote_data.sh $TIMESTAMP/remote
