#!/bin/bash

# VERSION="${VERSION:-"1.46.0"}"
ORG="${ORG:-myOrg}"
OAUTH_CLIENT_ID="${TS_OAUTH_CLIENT_ID:-$1}"
OAUTH_CLIENT_SECRET="${TS_OAUTH_CLIENT_SECRET:-$2}"

# Download Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# Retrieve Tailscale Access Token
if [[ -z $OAUTH_CLIENT_ID || -z $OAUTH_CLIENT_SECRET ]]; then
    exit 0
fi

ACCESS_TOKEN=$(curl \
    -d "client_id=${OAUTH_CLIENT_ID}" \
    -d "client_secret=${OAUTH_CLIENT_SECRET}" \
    "https://api.tailscale.com/api/v2/oauth/token" | jq -r '.access_token'
)

if [ $? == 0 ]; then
    echo "##[info] Retrieved 'access_token' from 'api/v2/oauth/token'"
else
    echo "##[error] Failed to retrieve 'access_token' from 'api/v2/oauth/token'"
    exit 1
fi

# Generate Auth Key
AUTH_TOKEN=$(curl \
    -u "$ACCESS_TOKEN:$OAUTH_CLIENT_SECRET" \
    --data-binary '
    {
        "capabilities": {
            "devices": {
                "create": {
                    "reusable": false,
                    "ephemeral": true,
                    "preauthorized": true,
                    "tags": [ "tag:ci-runner" ]
                }
            }
        },
        "expirySeconds": 1800
    }' \
    "https://api.tailscale.com/api/v2/tailnet/$ORG/keys" | jq -r '.key'
)

if [ $? == 0 ]; then
    echo "##[info] Retrieved 'authkey' from 'api/v2/tailnet/$ORG/keys'"
else
    echo "##[error] Failed to retrieve 'authkey' from 'api/v2/tailnet/$ORG/keys'"
    exit 1
fi

# Run Tailscaled
sudo tailscaled \
${TAILSCALED_ADDITIONAL_ARGS} 2>~/tailscaled.log &
if [ -z "${HOSTNAME}" ]; then
    HOSTNAME="github-$(cat /etc/hostname)"
fi

# Run Tailscale Up
sudo tailscale up --authkey ${AUTH_TOKEN} --hostname=${HOSTNAME} --accept-routes ${TAILSCALE_ADDITIONAL_ARGS}
if [ $? == 0 ]; then
    echo "##[info] Tailscale running."
else
    echo "##[error] Failed to start tailscale."
    exit 1
fi