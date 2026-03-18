#!/usr/bin/env bash

function one_line_pem {
        echo "$(awk 'NF {sub(/\n/, ""); printf "%s\\\\n",$0;}' $1)"
}

function json_ccp {
        local PP=$(one_line_pem $6)
        local CP=$(one_line_pem $7)
        sed -e "s/\${ORG_MSP}/$1/" \
                -e "s/\${DOMAIN}/$2/" \
                -e "s/\${P0PORT}/$3/" \
                -e "s/\${CAPORT}/$4/" \
                -e "s/\${P1PORT}/$5/" \
                -e "s#\${PEERPEM}#$PP#" \
                -e "s#\${CAPEM}#$CP#" \
                organizations/ccp-template.json
}

function yaml_ccp {
        local PP=$(one_line_pem $6)
        local CP=$(one_line_pem $7)
        sed -e "s/\${ORG_MSP}/$1/" \
                -e "s/\${DOMAIN}/$2/" \
                -e "s/\${P0PORT}/$3/" \
                -e "s/\${CAPORT}/$4/" \
                -e "s/\${P1PORT}/$5/" \
                -e "s#\${PEERPEM}#$PP#" \
                -e "s#\${CAPEM}#$CP#" \
                organizations/ccp-template.yaml | sed -e $'s/\\\\n/\\\n          /g'
}

ORG_MSP="Business"
DOMAIN="business.com"
P0PORT=7051
CAPORT=7054
P1PORT=7061
PEERPEM=organizations/peerOrganizations/${DOMAIN}/tlsca/tlsca.${DOMAIN}-cert.pem
CAPEM=organizations/peerOrganizations/${DOMAIN}/ca/ca.${DOMAIN}-cert.pem

echo "$(json_ccp $ORG_MSP $DOMAIN $P0PORT $CAPORT $P1PORT $PEERPEM $CAPEM)" >organizations/peerOrganizations/${DOMAIN}/connection-org1.json
echo "$(yaml_ccp $ORG_MSP $DOMAIN $P0PORT $CAPORT $P1PORT $PEERPEM $CAPEM)" >organizations/peerOrganizations/${DOMAIN}/connection-org1.yaml

ORG_MSP="Privacy"
DOMAIN="privacy.com"
P0PORT=9051
CAPORT=8054
P1PORT=9061
PEERPEM=organizations/peerOrganizations/${DOMAIN}/tlsca/tlsca.${DOMAIN}-cert.pem
CAPEM=organizations/peerOrganizations/${DOMAIN}/ca/ca.${DOMAIN}-cert.pem

echo "$(json_ccp $ORG_MSP $DOMAIN $P0PORT $CAPORT $P1PORT $PEERPEM $CAPEM)" >organizations/peerOrganizations/${DOMAIN}/connection-org2.json
echo "$(yaml_ccp $ORG_MSP $DOMAIN $P0PORT $CAPORT $P1PORT $PEERPEM $CAPEM)" >organizations/peerOrganizations/${DOMAIN}/connection-org2.yaml
