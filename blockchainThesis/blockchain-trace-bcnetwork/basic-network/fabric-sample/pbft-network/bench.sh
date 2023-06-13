#!/bin/bash

if [ $# -eq 0 ]; then
    echo "Usage: $0 <order_count>"
    exit 1
fi

if ! [[ "$1" =~ ^[0-9]+$ ]]; then
  echo "Please enter a valid number as the first argument."
  exit 1
fi

if (( $1 > 1024 || $1 % 3 != 1 )); then
  echo "Error: The first argument must be less than or equal to 1024, and leaves a remainder of 1 when divided by 3."
  exit 1
fi

benchmark_network="./benchmarks/network.yaml"
benchmark_config="./benchmarks/config.yaml"
peer_count=1
order_count=$1
peerlist1=""
peerlist2=""
orderlist=""
orders=""
peers=""
client_peers=""
channel_peers=""
containers=""

for i in $(seq 0 $(( ${order_count} - 1 ))); do
  port=$(expr 6050 + $i)
  containers+="    - orderer${i}.trace.com\n"
  orderlist+="      - orderer${i}.trace.com\n"
  orders+="
  orderer${i}.trace.com:
    url: grpc://localhost:${port}
    grpcOptions:
      grpc.keepalive_time_ms: 10000"
done

tmp_peer=5
for i in $(seq 0 $(( ${tmp_peer} - 1 ))); do
  containers+="    - peer${i}.orga.com\n"
done

for i in $(seq 0 $(( ${peer_count} - 1 ))); do
  port=$(expr 7051 + 100 \* $i)
  #containers+="    - peer${i}.orga.com\n"
  peerlist1+="      - peer${i}.orga.com\n"
  peerlist2+="    - peer${i}.orga.com\n"
  channel_peers+="
      peer${i}.orga.com:
        endorsingPeer: true
        chaincodeQuery: true
        ledgerQuery: true
        eventSource: true"
  client_peers+="
  peer${i}.orga.com:
    client:
      organization: OrgA
      credentialStore:
        path: /tmp/crypto/orga
        cryptoStore:
          path: /tmp/crypto/orga
      clientPrivateKey:
        path: crypto-config/peerOrganizations/orga.com/users/User1@orga.com/msp/keystore/key.pem
      clientSignedCert:
        path: crypto-config/peerOrganizations/orga.com/users/User1@orga.com/msp/signcerts/User1@orga.com-cert.pem"
  peers+="
  peer${i}.orga.com:
    url: grpc://localhost:${port}
    grpcOptions:
      grpc.keepalive_time_ms: 10000"
done


benchmark_network_yaml="name: Fabric
version: \"1.0\"

mutual-tls: false

caliper:
  blockchain: fabric
  command:
    start: scripts/gen.sh;scripts/utils.sh up
    end: scripts/utils.sh

info:
  Version: 1.4.4
  Size: ${order_count} Orders with ${tmp_peer} Peer
  Orderer: Pbft
  Distribution: Single Host(no Byz)
  StateDB: GoLevelDB

clients:${client_peers}

channels:
  mychannel:
    configBinary: ./channel-artifacts/channel.tx
    created: true
    orderers:
${orderlist}
    peers:${channel_peers}

    chaincodes:
    - id: money_demo
      version: \"1.0\"
      contractID: money_demo
      language: golang
      path: ../chaincode/demo
      targetPeers:
${peerlist1}

organizations:
  OrgA:
    mspid: OrgAMSP
    adminPrivateKey:
      path: crypto-config/peerOrganizations/orga.com/users/Admin@orga.com/msp/keystore/key.pem
    signedCert:
      path: crypto-config/peerOrganizations/orga.com/users/Admin@orga.com/msp/signcerts/Admin@orga.com-cert.pem
    peers:
${peerlist2}

orderers:${orders}

peers:${peers}\n"



benchmark_config_yaml="
test:
  name: pbft-network
  description: pbft-network
  workers:
    type: local
    number: $(expr ${order_count} + ${tmp_peer})

  rounds:
  - label: open
    description: open
    txNumber: 1000
    rateControl:
      type: fixed-rate
      opts:
        tps: 400
    callback: ../chaincode/demo/callback/open.js

  - label: query
    description: query
    txNumber: 1000
    rateControl:
      type: fixed-rate
      opts:
        tps: 400
    callback: ../chaincode/demo/callback/query.js

  - label: delete
    description: delete
    txNumber: 1000
    rateControl:
      type: fixed-rate
      opts:
        tps: 400
    callback: ../chaincode/demo/callback/delete.js

monitor:
  interval: 1
  type: 
    - docker
  docker:
    containers:
${containers}\n
"
echo "write benchmark_network_yaml yaml to ${benchmark_network}"
echo -e "${benchmark_network_yaml}" > ${benchmark_network}
echo "write ${benchmark_network} successfully!"
echo ""


echo "write benchmark_config_yaml yaml to ${benchmark_config}"
echo -e "${benchmark_config_yaml}" > ${benchmark_config}
echo "write ${benchmark_config} successfully!"
echo ""

