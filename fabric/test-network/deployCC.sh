#!/bin/bash

CC_NAME="authcc"
CC_SRC_PATH="../../chaincode-go"
CC_PKG="${CC_NAME}.tar.gz"
CC_LABEL="${CC_NAME}_1.0"
PEER_BIN="../bin/peer"
ORDERER_CA="${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt"

GO111MODULE=on GOWORK=off $PEER_BIN lifecycle chaincode package "$CC_PKG" --path "$CC_SRC_PATH" --lang golang --label "$CC_LABEL"

installCC() {
        local ORG_DOMAIN=$1
        local MSP_ID=$2
        local PEER_NAME=$3
        local PORT=$4
        set_peer_env "$ORG_DOMAIN" "$MSP_ID" "$PEER_NAME" "$PORT"

        $PEER_BIN lifecycle chaincode install "$CC_PKG"
}

installCC "business.com" "BusinessMSP" "peer0" 7051
installCC "business.com" "BusinessMSP" "peer1" 7061
installCC "privacy.com" "PrivacyMSP" "peer0" 9051
installCC "privacy.com" "PrivacyMSP" "peer1" 9061

PACKAGE_ID=$($PEER_BIN lifecycle chaincode queryinstalled | grep "$CC_LABEL" | cut -d ' ' -f 3 | tr -d ',')

# 4. 组织审批函数
approveCC() {
        local ORG_DOMAIN=$1
        local MSP_ID=$2
        local PORT=$3
        set_peer_env "$ORG_DOMAIN" "$MSP_ID" "peer0" "$PORT"

        $PEER_BIN lifecycle chaincode approveformyorg -o localhost:7050 \
                --ordererTLSHostnameOverride orderer.example.com \
                --tls --cafile "$ORDERER_CA" \
                --channelID "$CHANNEL_ID" --name "$CC_NAME" --version 1.0 \
                --package-id "$PACKAGE_ID" --sequence 1
}

approveCC "business.com" "BusinessMSP" 7051
approveCC "privacy.com" "PrivacyMSP" 9051

$PEER_BIN lifecycle chaincode commit -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls --cafile "$ORDERER_CA" \
        --channelID "$CHANNEL_ID" --name "$CC_NAME" --version 1.0 --sequence 1 \
        --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/business.com/peers/peer0.business.com/tls/ca.crt" \
        --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/privacy.com/peers/peer0.privacy.com/tls/ca.crt"
