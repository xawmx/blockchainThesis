#!/bin/bash

if [ $# -eq 0 ]; then
    echo "Usage: $0 <parameter>"
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


docker_compose_base="./base/docker-compose-base.yaml"
docker_compose_cli="docker-compose-cli.yaml"
crypto_config="crypto-config.yaml"
configtx="configtx.yaml"

docker_compose_base_yaml=""
docker_compose_cli_yaml=""
crypto_config_yaml=""
configtx_yaml=""

peer_count=5
order_count=$1
cli_orderer=""
cli_peer=""
cli_depend_list=""


configtx_orderer=""
configtx_peer=""
hostname=""

docker_compose_base_orderer=""
docker_compose_base_peer=""
nodeList=""

for i in $(seq 0 $(( ${order_count} - 1 ))); do
  port=$(expr 26070 + $i)
  nodeList+="http://orderer${i}.yzm.com:${port};"
done
nodeList=${nodeList::-1}
docker_compose_base_yaml="version: '2'\nservices:\n"

for i in $(seq 0 $(( ${order_count} - 1 ))); do
  port1=$(expr 6050 + $i)
  port2=$(expr 26070 + $i)
  docker_compose_base_orderer+="  orderer${i}.yzm.com:
    container_name: orderer${i}.trace.com
    extends:
      file: peer-base.yaml
      service: orderer-base
    environment:
        - ORDERER_GENERAL_LISTENPORT=${port1}
        - PBFT_LISTEN_PORT=${port2}
        - PBFT_NODE_ID=${i}
        - PBFT_NODE_TABLE=${nodeList}
    volumes:
        - ../channel-artifacts/genesis.block:/var/hyperledger/orderer/orderer.genesis.block
        - ../crypto-config/ordererOrganizations/yzm.com/orderers/orderer${i}.yzm.com/msp:/var/hyperledger/orderer/msp
        - ../crypto-config/ordererOrganizations/yzm.com/orderers/orderer${i}.yzm.com/tls:/var/hyperledger/orderer/tls
        - ../production/orderer:/var/hyperledger/production/orderer${i}
    ports:
      - ${port1}:${port1}
      - ${port2}:${port2}\n\n"


  # todo: ip
  cli_depend_list+="      - orderer${i}.yzm.com\n"
  ip="172.23.0."$(expr 100 + $i)
  cli_orderer+="
  orderer${i}.yzm.com:
    extends:
      file:   base/docker-compose-base.yaml
      service: orderer${i}.yzm.com
    container_name: orderer${i}.trace.com
    networks:
      solonet:
        ipv4_address: ${ip}\n"


  port=$(expr 6050 + $i)
  configtx_orderer+="        - orderer${i}.yzm.com:${port}\n"
  hostname+="      - Hostname: orderer${i}\n"
done

for i in $(seq 0 $(( ${peer_count} - 1 ))); do
  port1=$(expr 7051 + 100 \* ${i} )
  docker_compose_base_peer+="  peer${i}.orga.com:
    container_name: peer${i}.orga.com
    extends:
      file: peer-base.yaml
      service: peer-base
    environment:
      - CORE_PEER_ID=peer${i}.orga.com
      - CORE_PEER_ADDRESS=peer${i}.orga.com:7051
      - CORE_PEER_LISTENADDRESS=0.0.0.0:7051
      - CORE_PEER_CHAINCODEADDRESS=peer${i}.orga.com:7052
      - CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:7052
      - CORE_PEER_GOSSIP_BOOTSTRAP=peer${i}.orga.com:7051
      - CORE_PEER_GOSSIP_EXTERNALENDPOINT=peer${i}.orga.com:7051
      - CORE_PEER_LOCALMSPID=OrgAMSP
    volumes:
        - /var/run/:/host/var/run/
        - ../crypto-config/peerOrganizations/orga.com/peers/peer${i}.orga.com/msp:/etc/hyperledger/fabric/msp
        - ../crypto-config/peerOrganizations/orga.com/peers/peer${i}.orga.com/tls:/etc/hyperledger/fabric/tls
        - ../production/orga${i}:/var/hyperledger/production

    ports:
      - ${port1}:7051\n\n"



  cli_depend_list+="      - peer${i}.orga.com\n"
  ip="172.23.0."$(expr 2 + $i)
  cli_peer+="
  peer${i}.orga.com:
    container_name: peer${i}.orga.com
    extends:
      file:  base/docker-compose-base.yaml
      service: peer${i}.orga.com
    networks:
      solonet:
        ipv4_address: ${ip}\n"
  


  configtx_peer+="            - Host: peer${i}.orga.com
              Port: 7051\n"
done

docker_compose_base_yaml+=${docker_compose_base_orderer}
docker_compose_base_yaml+=${docker_compose_base_peer}

echo "write docker_compose_base yaml to ${docker_compose_base}"
echo -e "${docker_compose_base_yaml}" > ${docker_compose_base}
echo "write ${docker_compose_base} successfully!"
echo ""


docker_compose_cli_yaml="version: '2'

networks:
  solonet:
    ipam:
      config:
        - subnet: 172.23.0.0/24
          gateway: 172.23.0.1

services:${cli_orderer}

${cli_peer}

  cli:
    container_name: cli
    image: hyperledger/fabric-tools
    tty: true
    stdin_open: true
    environment:
      - SYS_CHANNEL=sys_channel
      - GOPATH=/opt/gopath
      - CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
      - FABRIC_LOGGING_SPEC=INFO
      - CORE_PEER_ID=cli
      - CORE_PEER_ADDRESS=peer0.orga.com:7051
      - CORE_PEER_LOCALMSPID=OrgAMSP
      - CORE_PEER_TLS_ENABLED=false
      - CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/orga.com/users/Admin@orga.com/msp
    working_dir: /opt/gopath/src/github.com/hyperledger/fabric/peer
    command: /bin/bash
    volumes:
        - /var/run/:/host/var/run/
        - ./../chaincode/:/opt/gopath/src/github.com/chaincode
        - ./crypto-config:/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/
        - ./scripts:/opt/gopath/src/github.com/hyperledger/fabric/peer/scripts/
        - ./channel-artifacts:/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts
    depends_on:
${cli_depend_list}
    networks:
      solonet:
        ipv4_address: 172.23.0.200\n\n"


echo "write docker_compose_cli yaml to ${docker_compose_cli}"
echo -e "${docker_compose_cli_yaml}" > ${docker_compose_cli}
echo "write ${docker_compose_cli} successfully!"
echo ""


configtx_yaml="---
Organizations:
    - &OrdererOrg
        Name: OrdererOrg
        ID: OrdererMSP
        MSPDir: crypto-config/ordererOrganizations/yzm.com/msp
        Policies:
            Readers:
                Type: Signature
                Rule: \"OR('OrdererMSP.member')\"
            Writers:
                Type: Signature
                Rule: \"OR('OrdererMSP.member')\"
            Admins:
                Type: Signature
                Rule: \"OR('OrdererMSP.admin')\"
    - &OrgA
        Name: OrgAMSP
        ID: OrgAMSP
        MSPDir: crypto-config/peerOrganizations/orga.com/msp
        Policies:
            Readers:
                Type: Signature
                Rule: \"OR('OrgAMSP.admin', 'OrgAMSP.peer', 'OrgAMSP.client')\"
            Writers:
                Type: Signature
                Rule: \"OR('OrgAMSP.admin', 'OrgAMSP.client')\"
            Admins:
                Type: Signature
                Rule: \"OR('OrgAMSP.admin')\"
        AnchorPeers:
${configtx_peer}

Capabilities:
    Channel: &ChannelCapabilities
        V1_4_3: true
        V1_3: false
        V1_1: false

    Orderer: &OrdererCapabilities
        V1_4_2: true
        V1_1: false

    Application: &ApplicationCapabilities
        V1_4_2: true
        V1_3: false
        V1_2: false
        V1_1: false

Application: &ApplicationDefaults
    Organizations:

    Policies:
        Readers:
            Type: ImplicitMeta
            Rule: \"ANY Readers\"
        Writers:
            Type: ImplicitMeta
            Rule: \"ANY Writers\"
        Admins:
            Type: ImplicitMeta
            Rule: \"MAJORITY Admins\"

    Capabilities:
        <<: *ApplicationCapabilities

Orderer: &OrdererDefaults
    OrdererType: pbft
    Addresses:
${configtx_orderer}

    BatchTimeout: 2s
    BatchSize:
        MaxMessageCount: 1000
        AbsoluteMaxBytes: 256 MB
        PreferredMaxBytes: 512 KB

    Organizations:
    Policies:
        Readers:
            Type: ImplicitMeta
            Rule: \"ANY Readers\"
        Writers:
            Type: ImplicitMeta
            Rule: \"ANY Writers\"
        Admins:
            Type: ImplicitMeta
            Rule: \"MAJORITY Admins\"
        BlockValidation:
            Type: ImplicitMeta
            Rule: \"ANY Writers\"

Channel: &ChannelDefaults
    Policies:
        Readers:
            Type: ImplicitMeta
            Rule: \"ANY Readers\"
        Writers:
            Type: ImplicitMeta
            Rule: \"ANY Writers\"
        Admins:
            Type: ImplicitMeta
            Rule: \"MAJORITY Admins\"

    Capabilities:
        <<: *ChannelCapabilities

Profiles:
    Genesis:
        <<: *ChannelDefaults
        Orderer:
            <<: *OrdererDefaults
            Organizations:
                - *OrdererOrg
            Capabilities:
                <<: *OrdererCapabilities
        Consortiums:
            SampleConsortium:
                Organizations:
                    - *OrgA
    Channel:
        Consortium: SampleConsortium
        <<: *ChannelDefaults
        Application:
            <<: *ApplicationDefaults
            Organizations:
                - *OrgA
            Capabilities:
                <<: *ApplicationCapabilities\n"


echo "write configtx yaml to ${configtx}"
echo -e "${configtx_yaml}" > ${configtx}
echo "write ${configtx} successfully!"
echo ""


crypto_config_yaml="OrdererOrgs:
  - Name: Orderer
    Domain: yzm.com
    EnableNodeOUs: true # 控制节点目录中是否生成配置文件
    Specs:
${hostname}

PeerOrgs:
  - Name: OrgA
    Domain: orga.com
    EnableNodeOUs: true
    Template:
      Count: ${peer_count}
    Users:
      Count: 1\n\n"


echo "write crypto_config yaml to ${crypto_config}"
echo -e "${crypto_config_yaml}" > ${crypto_config}
echo "write ${crypto_config} successfully!"
echo ""

./bench.sh ${order_count}
