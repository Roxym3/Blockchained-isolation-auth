# 1. 声明 Orderer 的 CA 证书路径（用于 TLS 通信）
export ORDERER_CA=${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt

# 2. 调用 IssueTicket 函数：签发一张 ID 为 TICKET_001 的票据
../bin/peer chaincode invoke -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls --cafile "$ORDERER_CA" \
        -C cross-domain-channel -n authcc \
        --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/business.com/peers/peer0.business.com/tls/ca.crt" \
        --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/privacy.com/peers/peer0.privacy.com/tls/ca.crt" \
        -c '{"function":"IssueTicket","Args":["TICKET_001","PrivacyMSP","Roxy3","1765000000"]}'

../bin/peer chaincode query -C cross-domain-channel -n authcc -c '{"Args":["VerifyTicket","TICKET_001"]}'

docker ps
