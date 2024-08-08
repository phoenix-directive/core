#!/bin/bash

OLD_VERSION=release/v2.11
UPGRADE_HEIGHT=35
CHAIN_ID=phoenix-1
CHAIN_HOME=$(pwd)/chain-upgrade-data
DENOM=uluna
SOFTWARE_UPGRADE_NAME="v2.12"
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

sleep 15

VALOPER_ADDR_1=$($OLD_BINARY q staking validators --output=json --home $CHAIN_HOME | jq .validators[0].operator_address -r)

# Stake and assert it is staked
echo "Delegate"
NO_ECHO=$($OLD_BINARY tx staking delegate $VALOPER_ADDR_1 100000000uluna --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --from wallet1 -y)
sleep 2
DELEGATIONS=$($OLD_BINARY query staking delegations $WALLET_ADDR_1 --home $CHAIN_HOME --output=json | jq ".delegation_responses | length")
if [[ "$DELEGATIONS" == "0" ]]; then
    echo "Delegation failed"
    exit 1
fi

# Unbond and assert the unbonding delegation
echo "Unbond"
NO_ECHO=$($OLD_BINARY tx staking unbond $VALOPER_ADDR_1 1000000uluna --keyring-backend test --chain-id $CHAIN_ID --home $CHAIN_HOME --from wallet1 -y)
sleep 2
UNBONDINGS=$($OLD_BINARY query staking unbonding-delegations $WALLET_ADDR_1 --home $CHAIN_HOME --output=json | jq ".unbonding_responses | length" )
if [[ "$UNBONDINGS" == "0" ]]; then
    echo "Unbonding failed"
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

DELEGATIONS=$($NEW_BINARY query staking delegations $WALLET_ADDR_1 --home $CHAIN_HOME --output=json | jq ".delegation_responses | length")
echo "DELEGATIONS $DELEGATIONS" 
if [[ "$DELEGATIONS" == "0" ]]; then
    echo "Delegations removed when upgrading"
fi

UNBONDINGS=$($NEW_BINARY query staking unbonding-delegations $WALLET_ADDR_1 --home $CHAIN_HOME --output=json | jq ".unbonding_responses | length")
echo "UNBONDINGS $UNBONDINGS" 
if [[ "$UNBONDINGS" == "0" ]]; then
    echo "Unbondings removed when upgrading"
fi

BALANCES=$($NEW_BINARY query bank balances $WALLET_ADDR_1 --home $CHAIN_HOME --output=json | jq ".balances | length")
echo "BALANCES $BALANCES" 
if [[ "$BALANCES" == "0" ]]; then
    echo "Balance removed when upgrading"
fi

COMISSION_RATE=$($NEW_BINARY query staking validator $VALOPER_ADDR_1 --home $CHAIN_HOME --output=json | jq ".commission.commission_rates.rate" -r)
echo "COMISSION_RATE $COMISSION_RATE"
if [[ "$COMISSION_RATE" != "0.050000000000000000" ]]; then
    echo "Commission rate not updated"
fi

MAX_COMISSION_RATE=$($NEW_BINARY query staking validator $VALOPER_ADDR_1 --home $CHAIN_HOME --output=json | jq ".commission.commission_rates.max_rate" -r)
echo "MAX_COMISSION_RATE $MAX_COMISSION_RATE"
if [[ "$MAX_COMISSION_RATE" != "0.050000000000000000" ]]; then
    echo "Max commission rate not updated"
fi

COMISSION_RATE_CHANGE=$($NEW_BINARY query staking validator $VALOPER_ADDR_1 --home $CHAIN_HOME --output=json | jq ".commission.commission_rates.max_change_rate" -r)
echo "COMISSION_RATE_CHANGE $COMISSION_RATE_CHANGE"
if [[ "$COMISSION_RATE_CHANGE" != "0.010000000000000000" ]]; then
    echo "Commission rate change not preserved"
fi
