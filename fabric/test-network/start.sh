#!/bin/bash
export PATH="${PWD}/../bin:${PWD}/$PATH"
export DOCKER_SOCK=/var/run/docker.sock

CHANNEL_ID="cross-domain-channel"
BLOCK_FILE="./channel-artifacts/${CHANNEL_ID}.block"
COMPOSE_FILES="-f compose/compose-test-net.yaml -f compose/docker/docker-compose-test-net.yaml"

ORGS=(
        "business.com:BusinessMSP:7051"
        "privacy.com:PrivacyMSP:9051"
)

set_peer_env() {
        local domain=$1
        local msp=$2
        local peer_name=$3
        local port=$4

        export FABRIC_CFG_PATH="${PWD}/../config"
        export CORE_PEER_TLS_ENABLED=true
        export CORE_PEER_LOCALMSPID="$msp"
        export CORE_PEER_MSPCONFIGPATH="${PWD}/organizations/peerOrganizations/${domain}/users/Admin@${domain}/msp"
        export CORE_PEER_TLS_ROOTCERT_FILE="${PWD}/organizations/peerOrganizations/${domain}/peers/${peer_name}.${domain}/tls/ca.crt"
        export CORE_PEER_ADDRESS="localhost:${port}"
}

export FABRIC_CFG_PATH=$PWD
for cfg in org1 org2 orderer; do
        cryptogen generate --config="./organizations/cryptogen/crypto-config-${cfg}.yaml" --output="organizations"
done

export FABRIC_CFG_PATH="${PWD}/configtx"
configtxgen -profile ChannelUsingRaft -outputBlock "$BLOCK_FILE" -channelID "$CHANNEL_ID"

export FABRIC_CFG_PATH=$PWD
docker compose $COMPOSE_FILES up -d

while ! nc -z localhost 7053; do
        sleep 0.5
done

sleep 3

OSN_TLS_CA="${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt"
ADMIN_CERT="${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt"
ADMIN_KEY="${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key"

../bin/osnadmin channel join \
        --channelID "$CHANNEL_ID" \
        --config-block "$BLOCK_FILE" \
        -o localhost:7053 \
        --ca-file "$OSN_TLS_CA" \
        --client-cert "$ADMIN_CERT" \
        --client-key "$ADMIN_KEY"

for org_info in "${ORGS[@]}"; do
        IFS=":" read -r domain msp base_port <<<"$org_info"

        set_peer_env "$domain" "$msp" "peer0" "$base_port"
        ../bin/peer channel join -b "$BLOCK_FILE"

        set_peer_env "$domain" "$msp" "peer1" $((base_port + 10))
        ../bin/peer channel join -b "$BLOCK_FILE"
done

export -f set_peer_env
export CHANNEL_ID
. ./deployCC.sh

docker ps
