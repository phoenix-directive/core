#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT=${REPO_ROOT:-$(git rev-parse --show-toplevel)}
TEST_DIR=${TEST_DIR:-$(mktemp -d "${TMPDIR:-/tmp}/core-v23-upgrade.XXXXXX")}
CHAIN_HOME=$TEST_DIR/node
OLD_VERSION=${OLD_VERSION:-v2.23.0}
OLD_BINARY=${OLD_BINARY:-$TEST_DIR/terrad_old}
NEW_BINARY=${NEW_BINARY:-$TEST_DIR/terrad_new}
CHAIN_ID=upgrade-v23-test
RPC_PORT=${RPC_PORT:-27657}
NODE="tcp://127.0.0.1:$RPC_PORT"
CONTRACT_PATH=$REPO_ROOT/integration-tests/src/contracts/counter.wasm
export GOCACHE=${GOCACHE:-${TMPDIR:-/tmp}/core-pd-go-build-cache}
NODE_PID=""
stop_node() {
  if [[ -n "$NODE_PID" ]]; then
    kill "$NODE_PID" 2>/dev/null || true
    wait "$NODE_PID" 2>/dev/null || true
    NODE_PID=""
  fi
}
trap stop_node EXIT
echo "Test data and logs: $TEST_DIR"
mkdir -p "$CHAIN_HOME"
if [[ ! -x "$OLD_BINARY" ]]; then
  mkdir -p "$TEST_DIR/old-source"
  git -C "$REPO_ROOT" archive "$OLD_VERSION" | tar -xf - -C "$TEST_DIR/old-source"
  (cd "$TEST_DIR/old-source" && go build -mod=readonly -o "$OLD_BINARY" ./cmd/terrad)
fi
if [[ ! -x "$NEW_BINARY" ]]; then
  (cd "$REPO_ROOT" && go build -mod=readonly -o "$NEW_BINARY" ./cmd/terrad)
fi
query() { "$BIN" query "$@" --home "$CHAIN_HOME" --node "$NODE" --output json; }
height() {
  "$BIN" status --home "$CHAIN_HOME" --node "$NODE" 2>/dev/null |
    jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height' 2>/dev/null
}
wait_height() {
  local target=$1 current i
  for ((i=0; i<120; i++)); do
    current=$(height || true)
    if [[ "$current" =~ ^[0-9]+$ ]] && ((current >= target)); then return; fi
    if ! kill -0 "$NODE_PID" 2>/dev/null; then
      tail -60 "$TEST_DIR/node.log"; return 1
    fi
    sleep 1
  done
  echo "Timed out waiting for height $target"; tail -60 "$TEST_DIR/node.log"; return 1
}
start_node() {
  "$BIN" start --home "$CHAIN_HOME" --pruning=nothing >> "$TEST_DIR/node.log" 2>&1 &
  NODE_PID=$!
}
tx() {
  local response hash code i
  response=$("$BIN" tx "$@" --from val1 --keyring-backend test --home "$CHAIN_HOME" \
    --node "$NODE" --chain-id "$CHAIN_ID" --gas auto --gas-adjustment 1.5 -y --output json)
  [[ $(jq -r '.code // 0' <<< "$response") == 0 ]] || { echo "$response"; return 1; }
  hash=$(jq -er .txhash <<< "$response")
  for ((i=0; i<30; i++)); do
    if TX_RESULT=$(query tx "$hash" 2>/dev/null); then
      code=$(jq -r '.tx_response.code // .code // 0' <<< "$TX_RESULT")
      [[ "$code" == 0 ]] || { echo "$TX_RESULT"; return 1; }
      printf '%s\n' "$TX_RESULT" > "$TEST_DIR/last-tx.json"
      return
    fi
    sleep 1
  done
  echo "Transaction $hash was not included"; return 1
}
event_value() {
  jq -er --arg key "$1" '[.. | objects | select(.key? == $key) | .value] | last' <<< "$TX_RESULT"
}

BIN=$OLD_BINARY
"$BIN" init upgrade-test --home "$CHAIN_HOME" --chain-id "$CHAIN_ID" > "$TEST_DIR/init.log" 2>&1
printf '%s\n' 'clock post desk civil pottery foster expand merit dash seminar song memory figure uniform spice circle try happy obvious trash crime hybrid hood cushion' |
  "$BIN" keys add val1 --home "$CHAIN_HOME" --recover --keyring-backend test > "$TEST_DIR/key.log" 2>&1
ADDRESS=$("$BIN" keys show val1 --home "$CHAIN_HOME" --keyring-backend test --address)
"$BIN" genesis add-genesis-account "$ADDRESS" 1000000000000uluna --home "$CHAIN_HOME"
"$BIN" genesis gentx val1 1000000000uluna --home "$CHAIN_HOME" --chain-id "$CHAIN_ID" \
  --keyring-backend test --commission-rate 0.05 --commission-max-rate 0.20 --commission-max-change-rate 0.01
"$BIN" genesis collect-gentxs --home "$CHAIN_HOME" > "$TEST_DIR/gentx.log" 2>&1
python3 - "$CHAIN_HOME" "$RPC_PORT" <<'PY'
import json, pathlib, sys
home=pathlib.Path(sys.argv[1]); port=int(sys.argv[2])
p=home/'config/genesis.json'; g=json.loads(p.read_text())
g['app_state']['gov']['params']['voting_period']='3s'
g['app_state']['gov']['params']['max_deposit_period']='3s'
p.write_text(json.dumps(g))
p=home/'config/config.toml'; s=p.read_text()
for a,b in [('timeout_commit = "5s"','timeout_commit = "1s"'),
            ('tcp://127.0.0.1:26657',f'tcp://127.0.0.1:{port}'),
            ('tcp://0.0.0.0:26656',f'tcp://127.0.0.1:{port-1}'),
            ('127.0.0.1:26658',f'127.0.0.1:{port+1}')]: s=s.replace(a,b)
p.write_text(s)
p=home/'config/app.toml'; s=p.read_text().replace('0.0.0.0:9090',f'127.0.0.1:{port+2}').replace('0.0.0.0:9091',f'127.0.0.1:{port+3}')
p.write_text(s)
PY
start_node
wait_height 3
echo "Create token and Wasm contract on v2.22"
tx tokenfactory create-denom upgrade-test
DENOM=$(event_value new_token_denom)
tx tokenfactory mint "1000000000$DENOM"
tx wasm store "$CONTRACT_PATH"
CODE_ID=$(event_value code_id)
tx wasm instantiate "$CODE_ID" '{"count":7}' --amount "100000000$DENOM" --label upgrade-counter --no-admin
CONTRACT=$(event_value _contract_address)
query wasm contract-state smart "$CONTRACT" "{\"get_count\":{\"addr\":\"$ADDRESS\"}}" | jq -e '.data.count == 7' >/dev/null
GOV_ADDRESS=$(query auth module-account gov | jq -er '.account.base_account.address // .account.value.address // .account.address')
UPGRADE_HEIGHT=$(($(height) + 20))
cat > "$TEST_DIR/proposal.json" <<EOF
{"messages":[{"@type":"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade","authority":"$GOV_ADDRESS","plan":{"name":"v2.23","height":"$UPGRADE_HEIGHT","info":""}}],"metadata":"","deposit":"550000000uluna","title":"Upgrade v23","summary":"Upgrade integration test"}
EOF
tx gov submit-proposal "$TEST_DIR/proposal.json"
PROPOSAL_ID=$(event_value proposal_id)
tx gov vote "$PROPOSAL_ID" yes
for ((i=0; i<30; i++)); do
  STATUS=$(query gov proposal "$PROPOSAL_ID" | jq -r '.proposal.status // .status')
  [[ "$STATUS" == PROPOSAL_STATUS_PASSED ]] && break
  [[ "$STATUS" != PROPOSAL_STATUS_FAILED && "$STATUS" != PROPOSAL_STATUS_REJECTED ]] || exit 1
  sleep 1
done
[[ "$STATUS" == PROPOSAL_STATUS_PASSED ]]
echo "Waiting for v2.23 upgrade at height $UPGRADE_HEIGHT"
wait_height "$((UPGRADE_HEIGHT-1))"
for ((i=0; i<60; i++)); do
  grep -q 'UPGRADE.*v2.23.*NEEDED' "$TEST_DIR/node.log" && break
  sleep 1
done
grep -q 'UPGRADE.*v2.23.*NEEDED' "$TEST_DIR/node.log"
stop_node
BIN=$NEW_BINARY
start_node
wait_height "$((UPGRADE_HEIGHT+3))"
echo "Verify migrated state and modules"
query bank balances "$CONTRACT" | jq -e --arg denom "$DENOM" '.balances[] | select(.denom == $denom) | .amount == "100000000"' >/dev/null
query wasm contract-state smart "$CONTRACT" "{\"get_count\":{\"addr\":\"$ADDRESS\"}}" | jq -e '.data.count == 7' >/dev/null
tx wasm execute "$CONTRACT" '{"increment":{}}'
query wasm contract-state smart "$CONTRACT" "{\"get_count\":{\"addr\":\"$ADDRESS\"}}" | jq -e '.data.count == 8' >/dev/null
tx tokenfactory mint "123$DENOM"
tx wasm store "$CONTRACT_PATH"
query ratelimit list-rate-limits | jq -e '(.rate_limits // []) | length == 0' >/dev/null
EXPECTED_VM=$(cd "$REPO_ROOT" && go list -m -f '{{.Version}}' github.com/CosmWasm/wasmvm/v3)
ACTUAL_VM=$("$BIN" query wasm libwasmvm-version)
[[ "$ACTUAL_VM" == "${EXPECTED_VM#v}" ]]
echo "v23 upgrade integration test passed: height $(height), libwasmvm $ACTUAL_VM"
