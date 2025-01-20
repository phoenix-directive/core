#!/bin/bash

OLD_VERSION=release/v2.13
UPGRADE_HEIGHT=35
CHAIN_ID=test-1
CHAIN_HOME=$(pwd)/chain-upgrade-data
CONTRACT_PATH=$(pwd)/src/contracts/counter.wasm
DENOM=uluna
SOFTWARE_UPGRADE_NAME="v2.14"
GOV_PERIOD="3s"

VAL_MNEMONIC_1="clock post desk civil pottery foster expand merit dash seminar song memory figure uniform spice circle try happy obvious trash crime hybrid hood cushion"
VAL_MNEMONIC_2="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"
WALLET_MNEMONIC_1="banner spread envelope side kite person disagree path silver will brother under couch edit food venture squirrel civil budget number acquire point work mass"

export OLD_BINARY=$CHAIN_HOME/terrad_old
export NEW_BINARY=$CHAIN_HOME/terrad_new

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
    git clone https://github.com/terra-money/core
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

$OLD_BINARY genesis add-genesis-account $($OLD_BINARY --home $CHAIN_HOME keys show val1 --keyring-backend test -a) 100000000000uluna --home $CHAIN_HOME

CURRENT_TIME=$(date +%s)
echo "Current time: $CURRENT_TIME"
$OLD_BINARY genesis add-genesis-account $($OLD_BINARY --home $CHAIN_HOME keys show wallet1 --keyring-backend test -a) 100000000000uluna --vesting-amount 200000000uluna --vesting-start-time $CURRENT_TIME --vesting-end-time $(($CURRENT_TIME + 10000)) --home $CHAIN_HOME

$OLD_BINARY genesis gentx val1 1000000000uluna --home $CHAIN_HOME --chain-id $CHAIN_ID --keyring-backend test --commission-max-rate 0.01 --commission-rate 0.01 --commission-max-change-rate 0.01
$OLD_BINARY genesis collect-gentxs --home $CHAIN_HOME

sed -i -e "s/\"max_deposit_period\": \"172800s\"/\"max_deposit_period\": \"$GOV_PERIOD\"/g" $CHAIN_HOME/config/genesis.json
sed -i -e "s/\"voting_period\": \"172800s\"/\"voting_period\": \"$GOV_PERIOD\"/g" $CHAIN_HOME/config/genesis.json

sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CHAIN_HOME/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CHAIN_HOME/config/config.toml
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' $CHAIN_HOME/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $CHAIN_HOME/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $CHAIN_HOME/config/app.toml

# run old node
echo "Starting old binary on a separate process"
if [[ "$OSTYPE" == "darwin"* ]]; then
    screen -L -dmS node1 $OLD_BINARY start --log_level trace --log_format json --home $CHAIN_HOME --pruning=nothing
else
    screen -L -Logfile $CHAIN_HOME/log-screen.log -dmS node1 $OLD_BINARY start --log_level trace --log_format json --home $CHAIN_HOME --pruning=nothing
fi

sleep 5

VALOPER_ADDR_1=$($OLD_BINARY q staking validators --output=json --home $CHAIN_HOME | jq .validators[0].operator_address -r)

# Create a new token
echo "Create a new token"
NO_ECHO=$($OLD_BINARY tx tokenfactory create-denom test --from wallet1 --keyring-backend test --gas auto --gas-adjustment 1.5 --home $CHAIN_HOME --chain-id $CHAIN_ID -y)
sleep 1
TOKEN_DENOM=$($OLD_BINARY query tokenfactory denoms-from-creator $WALLET_ADDR_1 --home $CHAIN_HOME --output=json | jq .denoms[0] -r)
echo "TOKEN_DENOM $TOKEN_DENOM"
echo "Mint token"
NO_ECHO=$($OLD_BINARY tx tokenfactory mint 1000000000$TOKEN_DENOM --from wallet1 --keyring-backend test --gas auto --gas-adjustment 1.5 --home $CHAIN_HOME --chain-id $CHAIN_ID -y)
sleep 1

# Upload a contract
echo "Upload a contract"
# NO_ECHO=$($OLD_BINARY tx wasm store $CONTRACT_PATH --from wallet1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME -y)
NO_ECHO=$($OLD_BINARY tx wasm store $CONTRACT_PATH --from wallet1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --gas auto --gas-adjustment 1.5 -y)
sleep 1

# Instantiate a contract
echo "Instantiate a contract"
NO_ECHO=$($OLD_BINARY tx wasm instantiate 1 '{"count":0}' --amount 100000000$TOKEN_DENOM --from wallet1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --label "counter" --no-admin -y)
sleep 1
CONTRACT_ADDRESS=$($OLD_BINARY query wasm list-contract-by-code 1 --output=json --home $CHAIN_HOME | jq .contracts[0] -r)
echo "CONTRACT_ADDRESS $CONTRACT_ADDRESS"
CONTRACT_BALANCE=$($OLD_BINARY query bank balances $CONTRACT_ADDRESS --home $CHAIN_HOME --output=json | jq ".balances[0].amount")
if [[ "$CONTRACT_BALANCE" == "100000000" ]]; then
    echo "Contract balance is less than 100000000"
    exit 1
fi


GOV_ADDRESS=$($OLD_BINARY query auth module-account gov --home $CHAIN_HOME --output json | jq .account.base_account.address -r)
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
  "summary": "Source Code Version https://github.com/terra-money/core"
}' > $CHAIN_HOME/software-upgrade.json

echo "Submit proposal"
NO_ECHO=$($OLD_BINARY tx gov submit-proposal $CHAIN_HOME/software-upgrade.json --from val1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME  -y)
sleep 2
echo "Vote"
NO_ECHO=$($OLD_BINARY tx gov vote 1 yes --from val1 --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME  -y)

## determine block_height to halt
while true; do
    BLOCK_HEIGHT=$($OLD_BINARY status --home $CHAIN_HOME | jq '.SyncInfo.latest_block_height' -r)
    if [ $BLOCK_HEIGHT = "$UPGRADE_HEIGHT" ]; then
        # assuming running only 1 terrad
        echo "BLOCK HEIGHT = $UPGRADE_HEIGHT REACHED, STOPPING OLD BINARY"
        pkill terrad_old
        break
    else
        STATUS=$($OLD_BINARY query gov proposal 1 --output=json --home $CHAIN_HOME | jq ".status" -r)
        echo "BLOCK_HEIGHT = $BLOCK_HEIGHT $STATUS"
        sleep 1
    fi
done
sleep 1

# run new binary
echo "Starting new binary"
if [[ "$OSTYPE" == "darwin"* ]]; then
    screen -L -dmS node1 $NEW_BINARY start --log_level trace --log_format json --home $CHAIN_HOME --pruning=nothing
else
    screen -L -Logfile $CHAIN_HOME/log-screen.log -dmS node1 $NEW_BINARY start --log_level trace --log_format json --home $CHAIN_HOME --pruning=nothing
fi
sleep 15

CONTRACT_BALANCE=$($NEW_BINARY query bank balances $CONTRACT_ADDRESS --home $CHAIN_HOME --output=json | jq ".balances | length")
if [[ "$CONTRACT_BALANCE" != "0" ]]; then
    echo "Contract balance is not burnt"
    exit 1
fi

sleep 10

echo "Upgrade successful"
# kill screen
screen -X -S node1 quit
