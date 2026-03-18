#!/bin/bash

# 配置区域
DOCKER_FILES=(
        "compose/compose-test-net.yaml"
        "compose/docker/docker-compose-test-net.yaml"
)
FILES_TO_REMOVE=(
        "./channel-artifacts/cross-domain-channel.block"
        "./authcc.tar.gz"
)
DIRS_TO_REMOVE=(
        "./organizations/peerOrganizations"
        "./organizations/ordererOrganizations"
)

# 转换 Compose 文件参数
COMPOSE_ARGS=$(printf -- "-f %s " "${DOCKER_FILES[@]}")

export DOCKER_SOCK=/var/run/docker.sock
docker compose $COMPOSE_ARGS down -v --remove-orphans
# 仅当文件存在时才尝试删除
rm -f "${FILES_TO_REMOVE[@]}"
rm -rf "${DIRS_TO_REMOVE[@]}"
