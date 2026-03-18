#.#!/bin/bash
export FABRIC_CFG_PATH="${PWD}/../config"
export PATH="${PWD}/../bin:${PATH}"
export CHANNEL_NAME="cross-domain-channel"
export ORDERER_CA="${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt"

updateAnchor() {
        ORG_PREFIX=$1 # business 或 privacy
        PORT=$2       # 7051 或 9051
        MSP_ID=$3     # BusinessMSP 或 PrivacyMSP
        PEER_HOST="peer0.${ORG_PREFIX}.com"

        echo ">>> 开始更新 ${MSP_ID} 的锚节点..."

        # 1. 设置身份环境变量
        export CORE_PEER_LOCALMSPID="${MSP_ID}"
        export CORE_PEER_TLS_ROOTCERT_FILE="${PWD}/organizations/peerOrganizations/${ORG_PREFIX}.com/peers/${PEER_HOST}/tls/ca.crt"
        export CORE_PEER_MSPCONFIGPATH="${PWD}/organizations/peerOrganizations/${ORG_PREFIX}.com/users/Admin@${ORG_PREFIX}.com/msp"
        export CORE_PEER_ADDRESS="localhost:${PORT}"

        # 2. 获取最新配置
        peer channel fetch config config_block.pb -o localhost:7050 -c $CHANNEL_NAME --tls --cafile "$ORDERER_CA"
        configtxlator proto_decode --input config_block.pb --type common.Block | jq .data.data[0].payload.data.config >config.json

        # 🌟 3. 黑科技：动态查找该组织在通道配置里的真实名称
        GROUP_NAME=$(jq -r '.channel_group.groups.Application.groups | to_entries[] | select(.value.values.MSP.value.config.name == "'${MSP_ID}'") | .key' config.json)

        if [ -z "$GROUP_NAME" ]; then
                echo "❌ 致命错误: 在通道配置中找不到 MSP ID 为 ${MSP_ID} 的组织！"
                return 1
        fi
        echo "-> 🔍 成功锁定组织真实代号: [${GROUP_NAME}]"

        # 4. 使用真实名称进行安全、精准的 JSON 注入
        jq '.channel_group.groups.Application.groups["'${GROUP_NAME}'"].values["AnchorPeers"] = {"mod_policy": "Admins","value":{"anchor_peers": [{"host": "'${PEER_HOST}'","port": '${PORT}'}]},"version": "0"}' config.json >modified_config.json

        # 5. 计算差值并包装
        configtxlator proto_encode --input config.json --type common.Config --output config.pb
        configtxlator proto_encode --input modified_config.json --type common.Config --output modified_config.pb
        configtxlator compute_update --channel_id $CHANNEL_NAME --original config.pb --updated modified_config.pb --output config_update.pb

        configtxlator proto_decode --input config_update.pb --type common.ConfigUpdate | jq '{payload:{header:{channel_header:{channel_id:"'$CHANNEL_NAME'", type:2}},data:{config_update: .}}}' >config_update_in_envelope.json
        configtxlator proto_encode --input config_update_in_envelope.json --type common.Envelope --output config_update_envelope.pb

        # 6. 直接提交！（修改自己的锚节点只需自己的 Admin 签名，peer channel update 默认会附带当前环境身份的签名）
        peer channel update -f config_update_envelope.pb -c $CHANNEL_NAME -o localhost:7050 --tls --cafile "$ORDERER_CA"

        echo "✅ ${MSP_ID} 锚节点更新成功！"
        echo "------------------------------------------------"
        rm -f config_block.pb config.json modified_config.json config.pb modified_config.pb config_update.pb config_update_in_envelope.json config_update_envelope.pb
}

# 依次执行更新
updateAnchor "business" 7051 "BusinessMSP"
updateAnchor "privacy" 9051 "PrivacyMSP"

echo "🎉 全网锚节点动态定位并更新完毕！"json config.pb modified_config.pb config_update.pb config_update_in_envelope.json config_update_envelope.pb
