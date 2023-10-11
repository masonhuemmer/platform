#!/bin/bash
#####################################################################################
# Retrieve Agent Pool ID and update 'IdleAgents' setting
#####################################################################################

ORGANIZATION="${ORGANIZATION:-$1}"
PAT_TOKEN="${PAT_TOKEN:-$2}"
AGENT_POOL_NAME="${AGENT_POOL_NAME:-$3}"
IDLE_AGENTS="${IDLE_AGENTS:-$4}"

echo "##[debug] GET Azure DevOps '_apis/distributedtask/pools'"
result=$(curl -sSL -u :$PAT_TOKEN --location "https://dev.azure.com/$ORGANIZATION/_apis/distributedtask/pools?poolName=$AGENT_POOL_NAME&api-version=7.0" \
--header 'Content-Type: application/json' \
)

pool_id=$(echo $result | jq -r '.value[].id')
echo "##[info] Agent Pool '$pool_id' found."

echo "##[debug] PATCH Azure DevOps '_apis/distributedtask/elasticpools/$pool_id'"
result=$(curl -sSL -u :$PAT_TOKEN --write-out "%{http_code}\n" --request PATCH \
--location "https://dev.azure.com/$ORGANIZATION/_apis/distributedtask/elasticpools/$pool_id?api-version=7.0" \
--header 'Content-Type: application/json' \
--data "{
    desiredIdle: $IDLE_AGENTS
}")

echo "##[info] Agent Pool '$AGENT_POOL_NAME' desiredIdle set to '$IDLE_AGENTS'"
