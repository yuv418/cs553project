#!/bin/sh

# This assumes that terraform has deployed everything

run_for_seed () {
    # https://stackoverflow.com/questions/4412238/what-is-the-cleanest-way-to-ssh-and-run-multiple-commands-in-bash
    # need to be root!
    ssh -i $SSH_KEY_PATH $WORLDGEN_IP << EOF 
        mkdir -p /etc/systemd/system/flappygo-worldgen.service.d/
        # Hard mode
        echo "[Service]\nEnvironment=\"STABLE_WORLD_SEED=${1}\"]n" > /etc/systemd/system/flappygo-worldgen.service.d/seed.conf
        systemctl daemon-reload 
        systemctl restart flappygo-worldgen
EOF

}

run_iteration() {
    cd client-automation

    # you must set the GAME_IP
    INPUT_CSV=input_seeds/${1}.csv GAME_URL=https://${CLIENT_IP} poetry run python -i src/input_simulator.py  --origin-to-force-quic-on=$ENGINE_IP:4433,$MUSIC_IP:4434 --ignore-certificate-errors-spki-list=$(cat certs/spki_hash.txt)

    cd ..
}

run_for_seed 8525333463046388971

run_iteration
run_iteration
run_iteration
run_iteration

run_for_seed 6977347407732442987

run_iteration
run_iteration
run_iteration
run_iteration
