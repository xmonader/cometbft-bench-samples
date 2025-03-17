#!/bin/bash

# This script demonstrates how to use transaction batching with the application.
# It creates accounts, funds them, and then performs batch transfers.

set -e

# Configuration
APP_HOST="localhost"
APP_PORT="26657"
BASE_URL="http://${APP_HOST}:${APP_PORT}"

# Function to encode a JSON payload to hex
encode_json() {
    echo -n "$1" | xxd -p | tr -d '\n'
}

# Function to send a transaction
send_tx() {
    local payload="$1"
    local encoded=$(encode_json "$payload")
    curl -s -X POST "${BASE_URL}/broadcast_tx_commit?tx=0x${encoded}"
}

# Function to wait for a transaction to be committed
wait_for_tx() {
    sleep 1
}

echo "Creating accounts..."

# Create account 1 (source account)
send_tx '{"type":"create_account","id":1,"initial_balance":10000}'
wait_for_tx

# Create accounts 2-11 (destination accounts)
for i in {2..11}; do
    send_tx "{\"type\":\"create_account\",\"id\":$i,\"initial_balance\":0}"
    wait_for_tx
done

echo "Accounts created successfully."

echo "Performing individual transfers..."

# Perform 10 individual transfers
start_time=$(date +%s.%N)

for i in {2..11}; do
    send_tx "{\"type\":\"transfer\",\"from\":1,\"to\":$i,\"amount\":10}"
    wait_for_tx
done

end_time=$(date +%s.%N)
individual_duration=$(echo "$end_time - $start_time" | bc)

echo "Individual transfers completed in $individual_duration seconds."

# Reset account balances
send_tx '{"type":"create_account","id":1,"initial_balance":10000}'
wait_for_tx

for i in {2..11}; do
    send_tx "{\"type\":\"create_account\",\"id\":$i,\"initial_balance\":0}"
    wait_for_tx
done

echo "Performing batch transfer..."

# Create a batch of 10 transfers with individual signatures
batch_payload='{"type":"batch","operations":['
for i in {2..11}; do
    if [ $i -ne 2 ]; then
        batch_payload+=','
    fi
    
    # Create the operation data for signing
    op_data="{\"from\":1,\"to\":$i,\"amount\":10}"
    
    # Sign the operation (in a real scenario, this would use the actual private key)
    # Here we're using a placeholder signature for demonstration
    signature="signature_for_operation_$i"
    
    # Add the signed operation to the batch
    batch_payload+="{\"type\":\"transfer\",\"from\":1,\"to\":$i,\"amount\":10,\"signature\":\"$signature\"}"
done
batch_payload+=']}'

echo "Batch payload:"
echo "$batch_payload" | jq .

# Perform the batch transfer
start_time=$(date +%s.%N)
send_tx "$batch_payload"
wait_for_tx
end_time=$(date +%s.%N)
batch_duration=$(echo "$end_time - $start_time" | bc)

echo "Batch transfer completed in $batch_duration seconds."

# Calculate the speedup
speedup=$(echo "$individual_duration / $batch_duration" | bc -l)
echo "Speedup: ${speedup}x"

echo "Checking account balances..."

# Check the balances of all accounts
for i in {1..11}; do
    balance=$(curl -s -X GET "${BASE_URL}/abci_query?path=\"/accounts/$i\"" | jq -r '.result.response.value' | base64 -d | jq -r '.balance')
    echo "Account $i balance: $balance"
done

echo "Done."
