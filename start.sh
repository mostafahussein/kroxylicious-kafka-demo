docker compose up -d openbao
sleep 30
export CONTAINER_NAME="openbao"
export VAULT_TOKEN="root"
export VAULT_ADDR=${VAULT_ADDR:-"http://127.0.0.1:8200"}
export TRANSIT_PATH="transit"
export KEY_NAME="KEK_kroxylicious"
export POLICY_NAME="kroxylicious_encryption_filter_policy"
export TOKEN_DISPLAY_NAME="kroxylicious record encryption"
export TOKEN_PERIOD="768h"

EXEC="docker exec -e VAULT_ADDR=$VAULT_ADDR -e VAULT_TOKEN=$VAULT_TOKEN $CONTAINER_NAME vault"

echo "Enabling transit secrets engine at path: $TRANSIT_PATH"
$EXEC secrets enable -path="$TRANSIT_PATH" transit

echo "Creating Transit key: $KEY_NAME"
$EXEC write -f "${TRANSIT_PATH}/keys/${KEY_NAME}"

echo "Creating Vault policy: $POLICY_NAME"
docker exec -i -e VAULT_ADDR=$VAULT_ADDR -e VAULT_TOKEN=$VAULT_TOKEN $CONTAINER_NAME sh -c \
"vault policy write $POLICY_NAME -" <<EOF
path "${TRANSIT_PATH}/keys/KEK_*" {
  capabilities = ["read"]
}
path "${TRANSIT_PATH}/datakey/plaintext/KEK_*" {
  capabilities = ["update"]
}
path "${TRANSIT_PATH}/decrypt/KEK_*" {
  capabilities = ["update"]
}
EOF

echo "Creating periodic Vault token for Kroxylicious filter"
$EXEC token create \
    -display-name="$TOKEN_DISPLAY_NAME" \
    -policy="$POLICY_NAME" \
    -period="$TOKEN_PERIOD" \
    -no-default-policy \
    -orphan \
    -format=json | jq -r '.auth.client_token' > kroxylicious_vault_token

docker compose up -d kafka kafka-ui kroxylicious
