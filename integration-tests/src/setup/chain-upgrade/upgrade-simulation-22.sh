#!/bin/bash

OLD_VERSION=release/v2.21
UPGRADE_HEIGHT=50
CHAIN_ID=blacklist-v222-test-1
CHAIN_HOME=$(pwd)/chain-upgrade-data
CONTRACT_PATH=$(pwd)/src/contracts/counter.wasm
DENOM=uluna
SOFTWARE_UPGRADE_NAME="v2.22"
GOV_PERIOD="10s"
EXPEDITED_GOV_PERIOD="5s"

VAL_MNEMONIC_1="clock post desk civil pottery foster expand merit dash seminar song memory figure uniform spice circle try happy obvious trash crime hybrid hood cushion"
BLACKLIST_TEST_MNEMONIC="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"
WALLET_MNEMONIC_1="banner spread envelope side kite person disagree path silver will brother under couch edit food venture squirrel civil budget number acquire point work mass"
WALLET_MNEMONIC_2="veteran try aware erosion drink dance decade comic dawn museum release episode original list ability owner size tuition surface ceiling depth seminar capable only"
WALLET_MNEMONIC_3="vacuum burst ordinary enact leaf rabbit gather lend left chase park action dish danger green jeans lucky dish mesh language collect acquire waste load"

export OLD_BINARY=$CHAIN_HOME/terrad_old
export NEW_BINARY=$CHAIN_HOME/terrad_new

cleanup_node() {
    if command -v screen > /dev/null; then
        screen -X -S node1 quit > /dev/null 2>&1 || true
    elif command -v tmux > /dev/null; then
        tmux kill-session -t node1 > /dev/null 2>&1 || true
    fi
    pkill -f "^$OLD_BINARY start " || true
    pkill -f "^$NEW_BINARY start " || true
}
trap cleanup_node EXIT

wait_for_new_tx() {
    local label=$1
    local broadcast=$2
    local hash result
    if ! echo "$broadcast" | jq -e '.code == 0 and (.txhash | length > 0)' > /dev/null; then
        echo "$label could not be broadcast: $broadcast"
        return 1
    fi
    hash=$(echo "$broadcast" | jq -r '.txhash')
    for ((attempt=0; attempt<20; attempt++)); do
        if result=$($NEW_BINARY query tx "$hash" --home "$CHAIN_HOME" -o json 2>/dev/null); then
            if echo "$result" | jq -e '.code == 0' > /dev/null; then
                return 0
            fi
            echo "$label failed on chain: $result"
            return 1
        fi
        sleep 1
    done
    echo "$label was not included after 20 seconds: $hash"
    return 1
}

rm -rf /tmp/terra
rm -r $CHAIN_HOME
mkdir $CHAIN_HOME
killall terrad_old
killall terrad_new

# install old binary
if ! command -v $OLD_BINARY &> /dev/null
then
    mkdir -p /tmp/terra
    cd /tmp/terra
    git clone https://github.com/phoenix-directive/core
    cd core
    git checkout $OLD_VERSION
    make build
    cp /tmp/terra/core/build/terrad $CHAIN_HOME/terrad_old
    cd $CHAIN_HOME
fi

# install new binary
if ! command -v $NEW_BINARY &> /dev/null
then
  cd ../..
  make build
  cp build/terrad $NEW_BINARY
fi

# init genesis
$OLD_BINARY init test --home $CHAIN_HOME --chain-id=$CHAIN_ID
echo $VAL_MNEMONIC_1 | $OLD_BINARY keys add val1 --home $CHAIN_HOME --recover --keyring-backend=test
VAL_ADDR_1=$($OLD_BINARY keys show val1 --home $CHAIN_HOME --keyring-backend=test --output=json | jq .address -r)

echo $WALLET_MNEMONIC_1 | $OLD_BINARY keys add wallet1 --home $CHAIN_HOME --recover --keyring-backend=test
WALLET_ADDR_1=$($OLD_BINARY keys show wallet1 --home $CHAIN_HOME --keyring-backend=test --output=json | jq .address -r)

echo $WALLET_MNEMONIC_2 | $OLD_BINARY keys add wallet2 --home $CHAIN_HOME --recover --keyring-backend=test
WALLET_ADDR_2=$($OLD_BINARY keys show wallet2 --home $CHAIN_HOME --keyring-backend=test --output=json | jq .address -r)

echo $WALLET_MNEMONIC_3 | $OLD_BINARY keys add wallet3 --home $CHAIN_HOME --recover --keyring-backend=test
WALLET_ADDR_3=$($OLD_BINARY keys show wallet3 --home $CHAIN_HOME --keyring-backend=test --output=json | jq .address -r)

echo "$BLACKLIST_TEST_MNEMONIC" | $OLD_BINARY keys add blacklist-test --home $CHAIN_HOME --recover --keyring-backend=test
BLACKLIST_TEST_ADDR=$($OLD_BINARY keys show blacklist-test --home $CHAIN_HOME --keyring-backend=test --output=json | jq .address -r)
if [[ "$BLACKLIST_TEST_ADDR" != "terra1a698u5rm2x6y50x5m3q37tnn0k6d4rjpfc8e7h" ]]; then
    echo "Unexpected blacklist test address: $BLACKLIST_TEST_ADDR"
    exit 1
fi

$OLD_BINARY genesis add-genesis-account $($OLD_BINARY --home $CHAIN_HOME keys show val1 --keyring-backend test -a) 100000000000uluna --home $CHAIN_HOME
$OLD_BINARY genesis add-genesis-account $BLACKLIST_TEST_ADDR 10000000uluna --home $CHAIN_HOME

CURRENT_TIME=$(date +%s)
echo "Current time: $CURRENT_TIME"
$OLD_BINARY genesis add-genesis-account $($OLD_BINARY --home $CHAIN_HOME keys show wallet1 --keyring-backend test -a) 100000000000uluna --vesting-amount 200000000uluna --vesting-start-time $CURRENT_TIME --vesting-end-time $(($CURRENT_TIME + 10000)) --home $CHAIN_HOME

$OLD_BINARY genesis gentx val1 1000000000uluna --home $CHAIN_HOME --chain-id $CHAIN_ID --keyring-backend test --commission-max-rate 0.01 --commission-rate 0.01 --commission-max-change-rate 0.01
$OLD_BINARY genesis collect-gentxs --home $CHAIN_HOME

sed -i -e "s/\"max_deposit_period\": \"172800s\"/\"max_deposit_period\": \"$GOV_PERIOD\"/g" $CHAIN_HOME/config/genesis.json
sed -i -e "s/\"voting_period\": \"172800s\"/\"voting_period\": \"$GOV_PERIOD\"/g" $CHAIN_HOME/config/genesis.json
sed -i -e "s/\"expedited_voting_period\": \"86400s\"/\"expedited_voting_period\": \"$EXPEDITED_GOV_PERIOD\"/g" $CHAIN_HOME/config/genesis.json

sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CHAIN_HOME/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CHAIN_HOME/config/config.toml
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' $CHAIN_HOME/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $CHAIN_HOME/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $CHAIN_HOME/config/app.toml

# run old node
echo "Starting old binary on a separate process"
if command -v screen > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        screen -L -dmS node1 $OLD_BINARY start --log_level trace --log_format json --home $CHAIN_HOME --pruning=nothing
    else
        screen -L -Logfile $CHAIN_HOME/log-screen.log -dmS node1 $OLD_BINARY start --log_level trace --log_format json --home $CHAIN_HOME --pruning=nothing
    fi
elif command -v tmux > /dev/null; then
    tmux new-session -d -s node1 "$OLD_BINARY start --log_level trace --log_format json --home '$CHAIN_HOME' --pruning=nothing >> '$CHAIN_HOME/log-screen.log' 2>&1"
else
    echo "screen or tmux is required to run the upgrade test"
    exit 1
fi

sleep 5

# The test account must be able to send before v2.22 activates the blacklist.
PRE_UPGRADE_TX=$($OLD_BINARY tx bank send blacklist-test "$WALLET_ADDR_1" 1uluna --from blacklist-test --fees 1000uluna --gas 200000 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --broadcast-mode sync -y -o json 2>&1)
if ! echo "$PRE_UPGRADE_TX" | jq -e '.code == 0' > /dev/null; then
    echo "Blacklist test account could not send before upgrade: $PRE_UPGRADE_TX"
    exit 1
fi
sleep 2

# Establish and exercise an authz grant before the upgrade. It must remain in
# state so the post-upgrade test reaches the blacklist check for its granter.
AUTHZ_GRANT_TX=$($OLD_BINARY tx authz grant "$WALLET_ADDR_1" send --spend-limit 100uluna --from blacklist-test --fees 1000uluna --gas 200000 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --broadcast-mode sync -y -o json 2>&1)
if ! echo "$AUTHZ_GRANT_TX" | jq -e '.code == 0' > /dev/null; then
    echo "Could not create pre-upgrade authz grant: $AUTHZ_GRANT_TX"
    exit 1
fi
sleep 2

$OLD_BINARY tx bank send "$BLACKLIST_TEST_ADDR" "$WALLET_ADDR_1" 1uluna --from blacklist-test --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --generate-only -o json > "$CHAIN_HOME/authz-send.json"
PRE_UPGRADE_AUTHZ_TX=$($OLD_BINARY tx authz exec "$CHAIN_HOME/authz-send.json" --from wallet1 --fees 1000uluna --gas 200000 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --broadcast-mode sync -y -o json 2>&1)
if ! echo "$PRE_UPGRADE_AUTHZ_TX" | jq -e '.code == 0' > /dev/null; then
    echo "Pre-upgrade authz grant could not execute: $PRE_UPGRADE_AUTHZ_TX"
    exit 1
fi
sleep 2

VALOPER_ADDR_1=$($OLD_BINARY q staking validators --output=json --home $CHAIN_HOME | jq .validators[0].operator_address -r)

echo "Create new periodic vesting account $WALLET_ADDR_2"
echo '{
 	"start_time": '$(date +%s)',
 	"periods":[
 		{
 			"coins": "10000000uluna",
 			"length_seconds": 10000
 		},
 		{
 			"coins": "10000000uluna",
 			"length_seconds": 10000
 		}
 	]
}' > $CHAIN_HOME/create-periodic-vesting-account.json
NO_ECHO=$($OLD_BINARY tx vesting create-periodic-vesting-account $WALLET_ADDR_2 $CHAIN_HOME/create-periodic-vesting-account.json --from val1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME -y -o json)
sleep 1
NO_ECHO=$($OLD_BINARY tx vesting create-periodic-vesting-account $WALLET_ADDR_3 $CHAIN_HOME/create-periodic-vesting-account.json --from val1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME -y -o json)
sleep 1

GOV_ADDRESS=$($OLD_BINARY query auth module-account gov --home $CHAIN_HOME --output json | jq -r '.account.value.address // .account.base_account.address')
if [[ -z "$GOV_ADDRESS" || "$GOV_ADDRESS" == "null" ]]; then
    echo "Could not resolve the governance module address"
    exit 1
fi
echo '{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority" : "'"$GOV_ADDRESS"'",
      "plan" : {
        "name": "'"$SOFTWARE_UPGRADE_NAME"'",
        "time": "0001-01-01T00:00:00Z",
        "height": "'"$UPGRADE_HEIGHT"'",
        "upgraded_client_state": null
      }
    }
  ],
  "metadata": "",
  "deposit": "550000000'$DENOM'",
  "title": "Upgrade to '$SOFTWARE_UPGRADE_NAME'",
  "summary": "Source Code Version https://github.com/phoenix-directive/core",
  "expedited": true
}' > $CHAIN_HOME/software-upgrade.json

echo "Submit proposal"
PROPOSAL_TX=$($OLD_BINARY tx gov submit-proposal $CHAIN_HOME/software-upgrade.json --from val1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME -y -o json 2>&1)
if ! echo "$PROPOSAL_TX" | jq -e '.code == 0' > /dev/null; then
    echo "Could not submit upgrade proposal: $PROPOSAL_TX"
    exit 1
fi
sleep 2
PROPOSAL_DETAILS=$($OLD_BINARY query gov proposal 1 --output=json --home $CHAIN_HOME 2>&1)
if ! echo "$PROPOSAL_DETAILS" | jq -e '.proposal.expedited == true' > /dev/null; then
    echo "Expedited upgrade proposal was not found on chain: $PROPOSAL_DETAILS"
    exit 1
fi
echo "Vote"
VOTE_TX=$($OLD_BINARY tx gov vote 1 yes --from val1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME -y -o json 2>&1)
if ! echo "$VOTE_TX" | jq -e '.code == 0' > /dev/null; then
    echo "Could not vote for upgrade proposal: $VOTE_TX"
    exit 1
fi

## determine block_height to halt
while true; do
    BLOCK_HEIGHT=$($OLD_BINARY status --home $CHAIN_HOME | jq -r '.sync_info.latest_block_height')
    if [[ "$BLOCK_HEIGHT" =~ ^[0-9]+$ ]] && (( BLOCK_HEIGHT >= UPGRADE_HEIGHT )); then
        # assuming running only 1 terrad
        echo "BLOCK HEIGHT = $UPGRADE_HEIGHT REACHED, STOPPING OLD BINARY"
        pkill terrad_old
        break
    else
        STATUS=$($OLD_BINARY query gov proposal 1 --output=json --home $CHAIN_HOME | jq -r '.proposal.status')
        echo "BLOCK_HEIGHT = $BLOCK_HEIGHT $STATUS"
        sleep 1
    fi
done
sleep 1

# run new binary
echo "Starting new binary"
LOG_FILE="$CHAIN_HOME/log-screen.log"

if command -v screen > /dev/null; then
    screen -dmS node1 bash -c \
      "$NEW_BINARY start --log_format json --home '$CHAIN_HOME' --pruning=nothing >> '$LOG_FILE' 2>&1"
else
    tmux new-session -d -s node1 "$NEW_BINARY start --log_format json --home '$CHAIN_HOME' --pruning=nothing >> '$LOG_FILE' 2>&1"
fi

sleep 15

echo "Upgrade successful"

# get new block height
NEW_BLOCK_HEIGHT=$($NEW_BINARY status --home $CHAIN_HOME | jq '.sync_info.latest_block_height' -r)
echo "NEW_BLOCK_HEIGHT $NEW_BLOCK_HEIGHT"

# assert that the new block height is greater than the upgrade height
if [ $NEW_BLOCK_HEIGHT -le $UPGRADE_HEIGHT ]; then
    echo "New block height is less than or equal to the upgrade height"
    exit 1
fi

# The same account is blocked by the v2.22 ante handler after the upgrade.
BLACKLISTED_TX=$($NEW_BINARY tx bank send blacklist-test "$WALLET_ADDR_1" 1uluna --from blacklist-test --fees 1000uluna --gas 200000 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --broadcast-mode sync -y -o json 2>&1)
if ! echo "$BLACKLISTED_TX" | jq -e '.code != 0 and (.raw_log | contains("is blacklisted"))' > /dev/null; then
    echo "Blacklist test account was not rejected by ante: $BLACKLISTED_TX"
    exit 1
fi

# A grant created before the upgrade cannot bypass the blacklist afterward.
AUTHZ_GRANTS=$($NEW_BINARY query authz grants "$BLACKLIST_TEST_ADDR" "$WALLET_ADDR_1" /cosmos.bank.v1beta1.MsgSend --home $CHAIN_HOME -o json 2>&1)
if ! echo "$AUTHZ_GRANTS" | jq -e '(.grants // []) | length > 0' > /dev/null; then
    echo "Pre-upgrade authz grant did not survive the upgrade: $AUTHZ_GRANTS"
    exit 1
fi
BLACKLISTED_AUTHZ_TX=$($NEW_BINARY tx authz exec "$CHAIN_HOME/authz-send.json" --from wallet1 --fees 1000uluna --gas 200000 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --broadcast-mode sync -y -o json 2>&1)
if ! echo "$BLACKLISTED_AUTHZ_TX" | jq -e '.code != 0 and (.raw_log | contains("is blacklisted"))' > /dev/null; then
    echo "Pre-upgrade authz grant bypassed the blacklist: $BLACKLISTED_AUTHZ_TX"
    exit 1
fi

echo "Performing some sanity checks"

# Create a new token
echo "Create a new token"
CREATE_DENOM_TX=$($NEW_BINARY tx tokenfactory create-denom test --from wallet1 --keyring-backend test --gas auto --gas-adjustment 1.5 --home $CHAIN_HOME --chain-id $CHAIN_ID -y -o json)
wait_for_new_tx "Create denom" "$CREATE_DENOM_TX" || exit 1
TOKEN_DENOM=$($NEW_BINARY query tokenfactory denoms-from-creator $WALLET_ADDR_1 --home $CHAIN_HOME --output=json | jq .denoms[0] -r)
echo "TOKEN_DENOM $TOKEN_DENOM"
echo "Mint token"
MINT_TX=$($NEW_BINARY tx tokenfactory mint 1000000000$TOKEN_DENOM --from wallet1 --keyring-backend test --gas auto --gas-adjustment 1.5 --home $CHAIN_HOME --chain-id $CHAIN_ID -y -o json)
wait_for_new_tx "Mint token" "$MINT_TX" || exit 1

# Upload a contract
echo "Upload a contract"
# NO_ECHO=$($OLD_BINARY tx wasm store $CONTRACT_PATH --from wallet1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME -y)
STORE_TX=$($NEW_BINARY tx wasm store $CONTRACT_PATH --from wallet1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --gas auto --gas-adjustment 1.5 -y -o json)
wait_for_new_tx "Store contract" "$STORE_TX" || exit 1

# Instantiate a contract
echo "Instantiate a contract"
INSTANTIATE_TX=$($NEW_BINARY tx wasm instantiate 1 '{"count":0}' --amount 100000000$TOKEN_DENOM --from wallet1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --label "counter" --no-admin -y -o json)
wait_for_new_tx "Instantiate contract" "$INSTANTIATE_TX" || exit 1
CONTRACT_ADDRESS=$($NEW_BINARY query wasm list-contract-by-code 1 --output=json --home $CHAIN_HOME | jq .contracts[0] -r)
echo "CONTRACT_ADDRESS $CONTRACT_ADDRESS"
CONTRACT_BALANCE=$($NEW_BINARY query bank balances $CONTRACT_ADDRESS --home $CHAIN_HOME --output=json | jq -r '.balances[0].amount')
if [[ "$CONTRACT_BALANCE" != "100000000" ]]; then
    echo "Unexpected contract balance: $CONTRACT_BALANCE"
    exit 1
fi

# Check period vesting account
echo "Check period vesting account"
PERIOD_VESTING_BALANCE_1=$($NEW_BINARY query bank spendable-balances $WALLET_ADDR_2 --home $CHAIN_HOME --output=json | jq ".balances[0].amount")
echo "PERIOD_VESTING_BALANCE_1 $PERIOD_VESTING_BALANCE_1"
sleep 5
PERIOD_VESTING_BALANCE_2=$($NEW_BINARY query bank spendable-balances $WALLET_ADDR_2 --home $CHAIN_HOME --output=json | jq ".balances[0].amount")
echo "PERIOD_VESTING_BALANCE_2 $PERIOD_VESTING_BALANCE_2"
if (( ${PERIOD_VESTING_BALANCE_2//\"} <= ${PERIOD_VESTING_BALANCE_1//\"} )); then
    echo "Period vesting account balance must be updated every block"
    exit 1
fi

# Check libwasmvm version
echo "Check libwasmvm version"
WASMVM_VERSION=$($NEW_BINARY query wasm libwasmvm-version --home "$CHAIN_HOME")

if [[ "$WASMVM_VERSION" != "2.2.9" ]]; then
    echo "Wrong libwasmvm version: $WASMVM_VERSION"
    exit 1
fi
