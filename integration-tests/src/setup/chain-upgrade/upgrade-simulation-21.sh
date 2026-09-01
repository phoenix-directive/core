#!/usr/bin/env bash

set -euo pipefail

OLD_VERSION=${OLD_VERSION:-release/v2.20}
OLD_REPO=${OLD_REPO:-https://github.com/phoenix-directive/core}
UPGRADE_HEIGHT=${UPGRADE_HEIGHT:-50}
CHAIN_ID=${CHAIN_ID:-test-1}
REPO_ROOT=${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}
CHAIN_HOME=${CHAIN_HOME:-$REPO_ROOT/chain-upgrade-data}
VAL2_NODE_HOME=${VAL2_NODE_HOME:-$CHAIN_HOME-val2-node}
CONTRACT_PATH=${CONTRACT_PATH:-$REPO_ROOT/integration-tests/src/contracts/counter.wasm}
DENOM=${DENOM:-uluna}
SOFTWARE_UPGRADE_NAME=${SOFTWARE_UPGRADE_NAME:-v2.21}
GOV_PERIOD=${GOV_PERIOD:-3s}
VAL1_STAKE=${VAL1_STAKE:-1000000000}
VAL2_STAKE=${VAL2_STAKE:-1000000}

VAL_MNEMONIC_1="clock post desk civil pottery foster expand merit dash seminar song memory figure uniform spice circle try happy obvious trash crime hybrid hood cushion"
VAL_MNEMONIC_2="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"
WALLET_MNEMONIC_1="banner spread envelope side kite person disagree path silver will brother under couch edit food venture squirrel civil budget number acquire point work mass"
WALLET_MNEMONIC_2="veteran try aware erosion drink dance decade comic dawn museum release episode original list ability owner size tuition surface ceiling depth seminar capable only"
WALLET_MNEMONIC_3="vacuum burst ordinary enact leaf rabbit gather lend left chase park action dish danger green jeans lucky dish mesh language collect acquire waste load"

export OLD_BINARY=$CHAIN_HOME/terrad_old
export NEW_BINARY=$CHAIN_HOME/terrad_new
export GOCACHE=${GOCACHE:-${TMPDIR:-/tmp}/core-pd-go-build-cache}
LAST_TXHASH=""
LAST_TX_RESULT=""

tx() {
  local bin=$1
  shift
  local response code txhash tx_result tx_code err_file
  err_file="$CHAIN_HOME/last-tx-stderr.log"
  if ! response=$("$bin" tx "$@" --keyring-backend test --chain-id "$CHAIN_ID" --home "$CHAIN_HOME" -y -o json 2>"$err_file"); then
    echo "transaction command failed: $*"
    echo "$response"
    cat "$err_file"
    exit 1
  fi
  if ! jq -e . >/dev/null 2>&1 <<<"$response"; then
    echo "transaction returned non-json output: $*"
    echo "$response"
    cat "$err_file"
    exit 1
  fi
  code=$(jq -r '(.code // 0) | tonumber' <<<"$response")
  if (( code != 0 )); then
    echo "transaction failed in CheckTx: $*"
    echo "$response" | jq -r '.raw_log // .log // .'
    exit 1
  fi

  txhash=$(jq -r '.txhash // .hash // empty' <<<"$response")
  LAST_TXHASH=$txhash
  LAST_TX_RESULT=$response
  sleep 2
  if [[ -n "$txhash" ]]; then
    tx_result=$("$bin" query tx "$txhash" --output=json --home "$CHAIN_HOME" 2>/dev/null || true)
    if jq -e . >/dev/null 2>&1 <<<"$tx_result"; then
      tx_code=$(jq -r '(.tx_response.code // .code // 0) | tonumber' <<<"$tx_result")
      if (( tx_code != 0 )); then
        echo "transaction failed in DeliverTx: $*"
        echo "$tx_result" | jq -r '.tx_response.raw_log // .raw_log // .tx_response.log // .log // .'
        exit 1
      fi
      LAST_TX_RESULT=$tx_result
    fi
  fi
}

wait_for_height() {
  local bin=$1
  local target=$2
  while true; do
    local height
    height=$("$bin" status --home "$CHAIN_HOME" | jq -r '.SyncInfo.latest_block_height // .sync_info.latest_block_height')
    if (( height >= target )); then
      break
    fi
    echo "waiting for height $target, current $height"
    sleep 1
  done
}

proposal_status() {
  local bin=$1
  local proposal_id=$2
  "$bin" query gov proposal "$proposal_id" --output=json --home "$CHAIN_HOME" | jq -r '.proposal.status // .status'
}

wait_for_proposal_passed() {
  local bin=$1
  local proposal_id=$2
  while true; do
    local status
    status=$(proposal_status "$bin" "$proposal_id")
    echo "proposal $proposal_id status: $status"
    if [[ "$status" == "PROPOSAL_STATUS_PASSED" || "$status" == "Passed" ]]; then
      break
    fi
    if [[ "$status" == "PROPOSAL_STATUS_REJECTED" || "$status" == "PROPOSAL_STATUS_FAILED" || "$status" == "Rejected" || "$status" == "Failed" ]]; then
      echo "proposal $proposal_id did not pass"
      exit 1
    fi
    sleep 1
  done
}

bonded_pool_tokens() {
  local bin=$1
  "$bin" query staking pool --output=json --home "$CHAIN_HOME" |
    jq -r '.pool.bonded_tokens // .bonded_tokens'
}

bonded_validator_tokens() {
  local bin=$1
  "$bin" query staking validators --output=json --home "$CHAIN_HOME" |
    jq -r '[.validators[] | select(.status == "BOND_STATUS_BONDED" or .status == "Bonded") | .tokens | tonumber] | add // 0'
}

validator_status() {
  local bin=$1
  local valoper=$2
  "$bin" query staking validator "$valoper" --output=json --home "$CHAIN_HOME" |
    jq -r '.validator.status // .status'
}

wait_for_validator_status() {
  local bin=$1
  local valoper=$2
  local expected=$3
  while true; do
    local status
    status=$(validator_status "$bin" "$valoper")
    echo "validator $valoper status: $status"
    if [[ "$status" == "$expected" ]]; then
      break
    fi
    sleep 1
  done
}

assert_bonded_validator_count() {
  local bin=$1
  local expected=$2
  local count
  count=$("$bin" query staking validators --output=json --home "$CHAIN_HOME" |
    jq -r '[.validators[] | select(.status == "BOND_STATUS_BONDED" or .status == "Bonded")] | length')
  echo "bonded validator count=$count"
  if (( count != expected )); then
    echo "expected $expected bonded validators, got $count"
    exit 1
  fi
}

assert_bonded_pool_mismatch() {
  local bin=$1
  local pool validators
  pool=$(bonded_pool_tokens "$bin")
  validators=$(bonded_validator_tokens "$bin")
  echo "bonded pool=$pool bonded validator tokens=$validators"
  if [[ "$pool" == "$validators" ]]; then
    echo "expected bonded pool mismatch before upgrade"
    exit 1
  fi
}

assert_bonded_pool_matches_validators() {
  local bin=$1
  local pool validators
  pool=$(bonded_pool_tokens "$bin")
  validators=$(bonded_validator_tokens "$bin")
  echo "bonded pool=$pool bonded validator tokens=$validators"
  if [[ "$pool" != "$validators" ]]; then
    echo "bonded pool does not match bonded validator tokens"
    exit 1
  fi
}

submit_and_pass_gov_msg() {
  local bin=$1
  local file=$2
  local from=$3
  local proposal_id

  tx "$bin" gov submit-proposal "$file" --from "$from"
  proposal_id=$("$bin" query gov proposals --output=json --home "$CHAIN_HOME" | jq -r '[(.proposals // [])[] | (.id // .proposal_id)] | last // empty')
  if [[ -z "$proposal_id" ]]; then
    echo "failed to determine proposal id after submitting $file"
    exit 1
  fi
  tx "$bin" gov vote "$proposal_id" yes --from val1
  wait_for_proposal_passed "$bin" "$proposal_id"
}

start_node() {
  local bin=$1
  echo "Starting $bin"
  if [[ "$OSTYPE" == "darwin"* ]]; then
    screen -L -dmS node1 "$bin" start --log_level trace --log_format json --home "$CHAIN_HOME" --pruning=nothing
  else
    screen -L -Logfile "$CHAIN_HOME/log-screen.log" -dmS node1 "$bin" start --log_level trace --log_format json --home "$CHAIN_HOME" --pruning=nothing
  fi
}

stop_node() {
  screen -X -S node1 quit >/dev/null 2>&1 || true
  pkill terrad_old >/dev/null 2>&1 || true
  pkill terrad_new >/dev/null 2>&1 || true
}

trap stop_node EXIT

rm -rf /tmp/terra "$CHAIN_HOME" "$VAL2_NODE_HOME"
mkdir -p "$CHAIN_HOME"
stop_node

if ! command -v "$OLD_BINARY" >/dev/null 2>&1; then
  mkdir -p /tmp/terra
  cd /tmp/terra
  git clone "$OLD_REPO" core
  cd core
  git checkout "$OLD_VERSION"
  make build
  cp /tmp/terra/core/build/terrad "$OLD_BINARY"
  cd "$CHAIN_HOME"
fi

if ! command -v "$NEW_BINARY" >/dev/null 2>&1; then
  cd "$REPO_ROOT"
  make build
  cp build/terrad "$NEW_BINARY"
fi

if [[ ! -f "$CONTRACT_PATH" ]]; then
  echo "contract wasm not found at $CONTRACT_PATH"
  exit 1
fi

cd "$CHAIN_HOME"

"$OLD_BINARY" init test --home "$CHAIN_HOME" --chain-id="$CHAIN_ID"

echo "$VAL_MNEMONIC_1" | "$OLD_BINARY" keys add val1 --home "$CHAIN_HOME" --recover --keyring-backend=test
echo "$VAL_MNEMONIC_2" | "$OLD_BINARY" keys add val2 --home "$CHAIN_HOME" --recover --keyring-backend=test
echo "$WALLET_MNEMONIC_1" | "$OLD_BINARY" keys add wallet1 --home "$CHAIN_HOME" --recover --keyring-backend=test
echo "$WALLET_MNEMONIC_2" | "$OLD_BINARY" keys add wallet2 --home "$CHAIN_HOME" --recover --keyring-backend=test
echo "$WALLET_MNEMONIC_3" | "$OLD_BINARY" keys add wallet3 --home "$CHAIN_HOME" --recover --keyring-backend=test

VAL_ADDR_1=$("$OLD_BINARY" keys show val1 --home "$CHAIN_HOME" --keyring-backend=test --output=json | jq -r .address)
VAL_ADDR_2=$("$OLD_BINARY" keys show val2 --home "$CHAIN_HOME" --keyring-backend=test --output=json | jq -r .address)
WALLET_ADDR_1=$("$OLD_BINARY" keys show wallet1 --home "$CHAIN_HOME" --keyring-backend=test --output=json | jq -r .address)
WALLET_ADDR_2=$("$OLD_BINARY" keys show wallet2 --home "$CHAIN_HOME" --keyring-backend=test --output=json | jq -r .address)
WALLET_ADDR_3=$("$OLD_BINARY" keys show wallet3 --home "$CHAIN_HOME" --keyring-backend=test --output=json | jq -r .address)

"$OLD_BINARY" genesis add-genesis-account "$VAL_ADDR_1" 100000000000uluna --home "$CHAIN_HOME"
"$OLD_BINARY" genesis add-genesis-account "$VAL_ADDR_2" 100000000000uluna --home "$CHAIN_HOME"

CURRENT_TIME=$(date +%s)
"$OLD_BINARY" genesis add-genesis-account "$WALLET_ADDR_1" 100000000000uluna \
  --vesting-amount 200000000uluna \
  --vesting-start-time "$CURRENT_TIME" \
  --vesting-end-time "$((CURRENT_TIME + 10000))" \
  --home "$CHAIN_HOME"

mkdir -p "$CHAIN_HOME/config/gentx"
mkdir -p "$VAL2_NODE_HOME"
"$OLD_BINARY" init val2 --home "$VAL2_NODE_HOME" --chain-id="$CHAIN_ID" >/dev/null
VAL2_CONS_PUBKEY=$("$OLD_BINARY" tendermint show-validator --home "$VAL2_NODE_HOME")
VAL2_NODE_ID=$("$OLD_BINARY" tendermint show-node-id --home "$VAL2_NODE_HOME")
"$OLD_BINARY" genesis gentx val1 "${VAL1_STAKE}uluna" --home "$CHAIN_HOME" --chain-id "$CHAIN_ID" --keyring-backend test --commission-max-rate 0.01 --commission-rate 0.01 --commission-max-change-rate 0.01 --min-self-delegation 1 --output-document "$CHAIN_HOME/config/gentx/gentx-val1.json"
"$OLD_BINARY" genesis gentx val2 "${VAL2_STAKE}uluna" --home "$CHAIN_HOME" --chain-id "$CHAIN_ID" --keyring-backend test --commission-max-rate 0.01 --commission-rate 0.01 --commission-max-change-rate 0.01 --min-self-delegation 1 --pubkey "$VAL2_CONS_PUBKEY" --node-id "$VAL2_NODE_ID" --output-document "$CHAIN_HOME/config/gentx/gentx-val2.json"
"$OLD_BINARY" genesis collect-gentxs --home "$CHAIN_HOME"

jq '
  .app_state.staking.params.max_validators = 2 |
  .app_state.staking.params.unbonding_time = "3s" |
  .app_state.alliance.params.reward_delay_time = "0s"
' "$CHAIN_HOME/config/genesis.json" > "$CHAIN_HOME/config/genesis.tmp" &&
  mv "$CHAIN_HOME/config/genesis.tmp" "$CHAIN_HOME/config/genesis.json"

sed -i -e "s/\"max_deposit_period\": \"172800s\"/\"max_deposit_period\": \"$GOV_PERIOD\"/g" "$CHAIN_HOME/config/genesis.json"
sed -i -e "s/\"voting_period\": \"172800s\"/\"voting_period\": \"$GOV_PERIOD\"/g" "$CHAIN_HOME/config/genesis.json"
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' "$CHAIN_HOME/config/config.toml"
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' "$CHAIN_HOME/config/config.toml"
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' "$CHAIN_HOME/config/config.toml"
sed -i -e 's/enable = false/enable = true/g' "$CHAIN_HOME/config/app.toml"
sed -i -e 's/swagger = false/swagger = true/g' "$CHAIN_HOME/config/app.toml"

start_node "$OLD_BINARY"
sleep 5

echo "Verify both validators are bonded"
assert_bonded_validator_count "$OLD_BINARY" 2

VALOPER_ADDR_1=$("$OLD_BINARY" q staking validators --output=json --home "$CHAIN_HOME" |
  jq -r --arg tokens "$VAL1_STAKE" '(.validators // [])[] | select(.tokens == $tokens) | .operator_address' |
  head -n 1)
if [[ -z "$VALOPER_ADDR_1" ]]; then
  echo "failed to find val1 validator operator address"
  exit 1
fi
VALOPER_ADDR_2=$("$OLD_BINARY" q staking validators --output=json --home "$CHAIN_HOME" |
  jq -r --arg tokens "$VAL2_STAKE" '(.validators // [])[] | select(.tokens == $tokens) | .operator_address' |
  head -n 1)
if [[ -z "$VALOPER_ADDR_2" ]]; then
  echo "failed to find val2 validator operator address"
  exit 1
fi

echo "Create periodic vesting accounts"
cat > "$CHAIN_HOME/create-periodic-vesting-account.json" <<EOF
{
  "start_time": $(date +%s),
  "periods": [
    {"coins": "10000000uluna", "length_seconds": 10000},
    {"coins": "10000000uluna", "length_seconds": 10000}
  ]
}
EOF
tx "$OLD_BINARY" vesting create-periodic-vesting-account "$WALLET_ADDR_2" "$CHAIN_HOME/create-periodic-vesting-account.json" --from val1
tx "$OLD_BINARY" vesting create-periodic-vesting-account "$WALLET_ADDR_3" "$CHAIN_HOME/create-periodic-vesting-account.json" --from val1

echo "Create tokenfactory denom and whitelist it as an Alliance asset"
tx "$OLD_BINARY" tokenfactory create-denom alliance --from wallet1 --gas auto --gas-adjustment 1.5
ALLIANCE_DENOM=$("$OLD_BINARY" query tokenfactory denoms-from-creator "$WALLET_ADDR_1" --home "$CHAIN_HOME" --output=json | jq -r .denoms[0])
tx "$OLD_BINARY" tokenfactory mint "1000000000$ALLIANCE_DENOM" --from wallet1 --gas auto --gas-adjustment 1.5

GOV_ADDRESS=$("$OLD_BINARY" query auth module-account gov --home "$CHAIN_HOME" --output json |
  jq -r '.account.base_account.address // .account.value.address // .account.address // empty')
if [[ -z "$GOV_ADDRESS" ]]; then
  echo "failed to determine gov module account address"
  exit 1
fi

cat > "$CHAIN_HOME/create-alliance.json" <<EOF
{
  "messages": [
    {
      "@type": "/alliance.alliance.MsgCreateAlliance",
      "authority": "$GOV_ADDRESS",
      "denom": "$ALLIANCE_DENOM",
      "reward_weight": "0.100000000000000000",
      "take_rate": "0.000000000000000000",
      "reward_change_rate": "1.000000000000000000",
      "reward_change_interval": "0s",
      "reward_weight_range": {
        "min": "0.100000000000000000",
        "max": "0.100000000000000000"
      }
    }
  ],
  "metadata": "",
  "deposit": "550000000$DENOM",
  "title": "Create Alliance asset",
  "summary": "Create Alliance asset for bonded pool regression test"
}
EOF
submit_and_pass_gov_msg "$OLD_BINARY" "$CHAIN_HOME/create-alliance.json" wallet1

echo "Delegate Alliance asset to val2 while it is bonded"
tx "$OLD_BINARY" alliance delegate "$VALOPER_ADDR_2" "100000000$ALLIANCE_DENOM" --from wallet1 --gas auto --gas-adjustment 1.5
sleep 3
assert_bonded_pool_matches_validators "$OLD_BINARY"

echo "Undelegate val2 self-delegation and wait for val2 to become unbonded"
tx "$OLD_BINARY" staking unbond "$VALOPER_ADDR_2" "${VAL2_STAKE}uluna" --from val2 --gas auto --gas-adjustment 1.5
wait_for_validator_status "$OLD_BINARY" "$VALOPER_ADDR_2" "BOND_STATUS_UNBONDED"

echo "Update Alliance reward weight to zero"
cat > "$CHAIN_HOME/update-alliance-zero-reward-weight.json" <<EOF
{
  "messages": [
    {
      "@type": "/alliance.alliance.MsgUpdateAlliance",
      "authority": "$GOV_ADDRESS",
      "denom": "$ALLIANCE_DENOM",
      "reward_weight": "0.000000000000000000",
      "take_rate": "0.000000000000000000",
      "reward_change_rate": "1.000000000000000000",
      "reward_change_interval": "1s",
      "reward_weight_range": {
        "min": "0.000000000000000000",
        "max": "0.000000000000000000"
      }
    }
  ],
  "metadata": "",
  "deposit": "550000000$DENOM",
  "title": "Zero Alliance reward weight",
  "summary": "Set Alliance reward weight to zero before validating bonded pool accounting"
}
EOF
submit_and_pass_gov_msg "$OLD_BINARY" "$CHAIN_HOME/update-alliance-zero-reward-weight.json" wallet1
sleep 3

echo "Verifying old chain has bonded pool mismatch before upgrade"
assert_bonded_pool_mismatch "$OLD_BINARY"

cat > "$CHAIN_HOME/software-upgrade.json" <<EOF
{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "$GOV_ADDRESS",
      "plan": {
        "name": "$SOFTWARE_UPGRADE_NAME",
        "time": "0001-01-01T00:00:00Z",
        "height": "$UPGRADE_HEIGHT",
        "upgraded_client_state": null
      }
    }
  ],
  "metadata": "",
  "deposit": "550000000$DENOM",
  "title": "Upgrade to $SOFTWARE_UPGRADE_NAME",
  "summary": "Upgrade to $SOFTWARE_UPGRADE_NAME"
}
EOF

submit_and_pass_gov_msg "$OLD_BINARY" "$CHAIN_HOME/software-upgrade.json" val1

while true; do
  BLOCK_HEIGHT=$("$OLD_BINARY" status --home "$CHAIN_HOME" | jq -r '.SyncInfo.latest_block_height // .sync_info.latest_block_height')
  if (( BLOCK_HEIGHT >= UPGRADE_HEIGHT )); then
    echo "BLOCK HEIGHT = $UPGRADE_HEIGHT REACHED, STOPPING OLD BINARY"
    stop_node
    break
  fi
  echo "BLOCK_HEIGHT = $BLOCK_HEIGHT"
  sleep 1
done
sleep 1

start_node "$NEW_BINARY"
sleep 15

echo "Upgrade successful"
NEW_BLOCK_HEIGHT=$("$NEW_BINARY" status --home "$CHAIN_HOME" | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height')
echo "NEW_BLOCK_HEIGHT $NEW_BLOCK_HEIGHT"
if (( NEW_BLOCK_HEIGHT <= UPGRADE_HEIGHT )); then
  echo "New block height is less than or equal to the upgrade height"
  exit 1
fi

echo "Verifying migration repaired the bonded pool mismatch"
assert_bonded_pool_matches_validators "$NEW_BINARY"

echo "Performing some sanity checks"
tx "$NEW_BINARY" tokenfactory create-denom test --from wallet1 --gas auto --gas-adjustment 1.5
TOKEN_DENOM=$("$NEW_BINARY" query tokenfactory denoms-from-creator "$WALLET_ADDR_1" --home "$CHAIN_HOME" --output=json | jq -r '.denoms[-1]')
echo "TOKEN_DENOM $TOKEN_DENOM"
tx "$NEW_BINARY" tokenfactory mint "1000000000$TOKEN_DENOM" --from wallet1 --gas auto --gas-adjustment 1.5

tx "$NEW_BINARY" wasm store "$CONTRACT_PATH" --from wallet1 --gas auto --gas-adjustment 1.5
CODE_ID=$(jq -r '[.. | objects | select(has("key") and .key == "code_id") | .value] | last // empty' <<<"$LAST_TX_RESULT")
if [[ -z "$CODE_ID" ]]; then
  echo "failed to determine wasm code id from store transaction"
  echo "$LAST_TX_RESULT" | jq .
  exit 1
fi
tx "$NEW_BINARY" wasm instantiate "$CODE_ID" '{"count":0}' --amount "100000000$TOKEN_DENOM" --from wallet1 --label "counter" --no-admin
CONTRACT_ADDRESS=$("$NEW_BINARY" query wasm list-contract-by-code "$CODE_ID" --output=json --home "$CHAIN_HOME" | jq -r .contracts[0])
echo "CONTRACT_ADDRESS $CONTRACT_ADDRESS"
CONTRACT_BALANCE=$("$NEW_BINARY" query bank balances "$CONTRACT_ADDRESS" --home "$CHAIN_HOME" --output=json | jq -r '.balances[0].amount')
if [[ "$CONTRACT_BALANCE" != "100000000" ]]; then
  echo "Contract balance should be 100000000, got $CONTRACT_BALANCE"
  exit 1
fi

PERIOD_VESTING_BALANCE_1=$("$NEW_BINARY" query bank spendable-balances "$WALLET_ADDR_2" --home "$CHAIN_HOME" --output=json | jq -r ".balances[0].amount")
echo "PERIOD_VESTING_BALANCE_1 $PERIOD_VESTING_BALANCE_1"
sleep 5
PERIOD_VESTING_BALANCE_2=$("$NEW_BINARY" query bank spendable-balances "$WALLET_ADDR_2" --home "$CHAIN_HOME" --output=json | jq -r ".balances[0].amount")
echo "PERIOD_VESTING_BALANCE_2 $PERIOD_VESTING_BALANCE_2"
if (( PERIOD_VESTING_BALANCE_2 <= PERIOD_VESTING_BALANCE_1 )); then
  echo "Period vesting account balance must be updated every block"
  exit 1
fi

stop_node
echo "v21 Migration upgrade test passed"
