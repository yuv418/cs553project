#!/bin/sh

# This assumes that terraform has deployed everything

# setup env vars
OUTPUT=$(terraform -chdir=terraform output -json service_endpoints)
CLIENT_HOST=$(echo $OUTPUT | jq -r .client)
ENGINE_HOST=$(echo $OUTPUT | jq -r .engine)
MUSIC_HOST=$(echo $OUTPUT | jq -r .music)
INITIATOR_HOST=$(echo $OUTPUT | jq -r .initiator)
WORLDGEN_HOST=$(echo $OUTPUT | jq -r .worldgen)
SPKI_HASH=$(cat ./certs/spki_hash.txt)
DEPLOY_TIME=$(cat ./terraform/deploy_time.txt)
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

run_for_seed() {
    # https://stackoverflow.com/questions/4412238/what-is-the-cleanest-way-to-ssh-and-run-multiple-commands-in-bash
    # need to be root!
    ssh -i terraform/certs/ssh_key ec2-user@$WORLDGEN_HOST <<EOF
        sudo mkdir -p /etc/systemd/system/flappygo-worldgen.service.d/
        # Hard mode
        sudo sh -c 'echo -e "[Service]\nEnvironment=\"STABLE_WORLD_SEED=${1}\"" > /etc/systemd/system/flappygo-worldgen.service.d/seed.conf'
        sudo systemctl daemon-reload
        sudo systemctl restart flappygo-worldgen
EOF

}

run_iteration() {
    cd client-automation

    # you must set the GAME_HOST
    STAT_DIR=../stat/$DEPLOY_TIME/$TIMESTAMP/client_seed_${1}_run_${2} INPUT_CSV=input_seeds/${1}.csv GAME_URL=https://${CLIENT_HOST} poetry run python src/input_simulator.py --origin-to-force-quic-on=$ENGINE_HOST:4433,$MUSIC_HOST:4433 --ignore-certificate-errors-spki-list=${SPKI_HASH}

    cd ..

}

# https://stackoverflow.com/questions/49110/how-do-i-write-a-for-loop-in-bash
run_test() {
    run_for_seed $1
    for i in $(seq 1 5); do
        run_iteration $1 $i
    done

}

run_test 8525333463046388971
run_test 6977347407732442987

# done, now collect stats over all the runs
./collect_remote_data.sh $TIMESTAMP/remote
