#!/usr/bin/env bash
# =============================================================================
# demo-up.sh, bring up the FULL clickable demo stack and HOLD it.
#
# Unlike fork-demo.sh (which runs a scripted cast scenario then tears down),
# this stands the stack up and KEEPS IT RUNNING so you drive everything by hand
# from the UI:
#   anvil mainnet fork (chain id 31337)
#     -> DeployFork diamond (real Aave + Morpho + Pendle)
#     -> fund demo users with real USDC (impersonation)
#     -> write the diamond address into the UI's .env.local
#     -> start the keeper (composite live risk, signs as curator)
#     -> WAIT (Ctrl-C tears everything down)
#
# Then, in a second terminal:  cd vault-router-ui && npm run dev
# and open http://localhost:3000 .
#
# Usage:   ./scripts/demo-up.sh [fork-rpc-url]
# =============================================================================
set -euo pipefail

# FOUNDRY_BIN: directory holding anvil/forge/cast (default: where forge is on PATH)
# DIAMOND_REPO: path to a vault-router-diamond checkout (required)
# UI_REPO: path to a vault-router-ui checkout (required)
FOUNDRY="${FOUNDRY_BIN:-$(dirname "$(command -v forge)")}"
DIAMOND_REPO="${DIAMOND_REPO:?set DIAMOND_REPO to a vault-router-diamond checkout}"
KEEPER_REPO="$(cd "$(dirname "$0")/.." && pwd)"
UI_REPO="${UI_REPO:?set UI_REPO to a vault-router-ui checkout}"

FORK_URL="${1:-https://ethereum-rpc.publicnode.com}"
PORT=8549
RPC="http://127.0.0.1:${PORT}"
CHAIN_ID=31337

USDC=0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
# aEthUSDC reserve, impersonated as the USDC faucet (fork-only; touches no key).
USDC_WHALE=0x98C23E9d8f34FEFb1B7BD6a91B7FF122F4e16F5c

# anvil account #0 = diamond owner AND curator (DeployFork default).
DEPLOYER_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
# anvil accounts #1..#3, the demo users to import into the browser wallet.
ALICE=0x70997970C51812dc3A010C7d01b50e0d17dc79C8
BOB=0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
CAROL=0x90F79bf6EB2c4f870365E785982E1f101E93b906

cast() { "$FOUNDRY/cast" "$@"; }
log()  { printf '\n\033[1;36m== %s ==\033[0m\n' "$*"; }

PIDS=()
cleanup() { echo; echo "tearing down demo stack…"; for p in "${PIDS[@]:-}"; do kill "$p" 2>/dev/null || true; done; }
trap cleanup EXIT INT TERM

# --- 1. anvil mainnet fork (chain id 31337) ----------------------------------
log "starting anvil fork of Ethereum mainnet on :$PORT (chain id $CHAIN_ID)"
"$FOUNDRY/anvil" --fork-url "$FORK_URL" --port "$PORT" --chain-id "$CHAIN_ID" --silent &
PIDS+=($!)
for _ in $(seq 1 30); do cast chain-id --rpc-url "$RPC" >/dev/null 2>&1 && break; sleep 1; done
cast chain-id --rpc-url "$RPC" >/dev/null || { echo "anvil not ready"; exit 1; }

# --- 2. deploy the real-facet diamond ----------------------------------------
log "deploying diamond via DeployFork.s.sol (Aave + Morpho + Pendle, + Fee/Lock facets)"
[ -f "$DIAMOND_REPO/script/DeployFork.s.sol" ] || { echo "DeployFork.s.sol missing"; exit 1; }
(cd "$DIAMOND_REPO" && "$FOUNDRY/forge" script script/DeployFork.s.sol:DeployFork \
  --rpc-url "$RPC" --broadcast --private-key "$DEPLOYER_KEY" -vv)

VAULT=$(python3 - "$DIAMOND_REPO/broadcast/DeployFork.s.sol/$CHAIN_ID/run-latest.json" <<'EOF'
import json, sys
txs = json.load(open(sys.argv[1]))["transactions"]
print(next(t["contractAddress"] for t in txs
           if t.get("transactionType") == "CREATE" and t.get("contractName") == "Vault"))
EOF
)
log "diamond (vault) at $VAULT"

# Earliest block in the DeployFork broadcast = the diamond's deploy block.
# The UI floors the Activity feed's getLogs fromBlock here so it never scans
# past the fork tip (which would make anvil proxy upstream and return nothing).
DEPLOY_BLOCK=$(python3 - "$DIAMOND_REPO/broadcast/DeployFork.s.sol/$CHAIN_ID/run-latest.json" <<'EOF'
import json, sys
receipts = json.load(open(sys.argv[1]))["receipts"]
print(min(int(r["blockNumber"], 16) for r in receipts))
EOF
)
log "diamond deploy block: $DEPLOY_BLOCK"

# --- 3. fund the demo users with real USDC -----------------------------------
log "funding demo users with real USDC (impersonated aUSDC reserve)"
cast rpc anvil_impersonateAccount "$USDC_WHALE" --rpc-url "$RPC" >/dev/null
cast rpc anvil_setBalance "$USDC_WHALE" 0xDE0B6B3A7640000 --rpc-url "$RPC" >/dev/null
for u in "$ALICE:250000" "$BOB:100000" "$CAROL:50000"; do
  addr="${u%%:*}"; amt="${u##*:}"
  cast send "$USDC" "transfer(address,uint256)(bool)" "$addr" "${amt}000000" \
    --from "$USDC_WHALE" --unlocked --rpc-url "$RPC" >/dev/null
  echo "  $addr <- $amt USDC"
done
cast rpc anvil_stopImpersonatingAccount "$USDC_WHALE" --rpc-url "$RPC" >/dev/null

# --- 4. wire the UI: write the diamond address into .env.local ---------------
log "writing NEXT_PUBLIC_VAULT_ADDRESS + NEXT_PUBLIC_DEPLOY_BLOCK into the UI .env.local"
python3 - "$UI_REPO/.env.local" "$VAULT" "$DEPLOY_BLOCK" <<'EOF'
import sys, os
path, addr, deploy_block = sys.argv[1], sys.argv[2], sys.argv[3]
lines = open(path).read().splitlines() if os.path.exists(path) else []
out, seen_addr, seen_block = [], False, False
for ln in lines:
    if ln.startswith("NEXT_PUBLIC_VAULT_ADDRESS="):
        out.append(f"NEXT_PUBLIC_VAULT_ADDRESS={addr}"); seen_addr = True
    elif ln.startswith("NEXT_PUBLIC_DEPLOY_BLOCK="):
        out.append(f"NEXT_PUBLIC_DEPLOY_BLOCK={deploy_block}"); seen_block = True
    else:
        out.append(ln)
if not seen_addr:
    out += ["NEXT_PUBLIC_CHAIN_ID=31337", f"NEXT_PUBLIC_VAULT_ADDRESS={addr}",
            "NEXT_PUBLIC_FORK_RPC_URL=http://127.0.0.1:8549"]
if not seen_block:
    out.append(f"NEXT_PUBLIC_DEPLOY_BLOCK={deploy_block}")
open(path, "w").write("\n".join(out) + "\n")
print("  ->", path, "vault", addr, "deploy_block", deploy_block)
EOF

# --- 5. start the keeper (composite live risk, signs as curator) -------------
log "building + starting keeper (composite risk, live signing)"
(cd "$KEEPER_REPO" && GOWORK=off go build -o bin/keeper ./cmd/keeper)
set -a; . "$KEEPER_REPO/.env.fork"; set +a
export KEEPER_RPC_URL="$RPC" KEEPER_VAULT_ADDRESS="$VAULT"
"$KEEPER_REPO/bin/keeper" 2>&1 | sed 's/^/  [keeper] /' &
PIDS+=($!)

# --- 6. hold ------------------------------------------------------------------
cat <<EOF

\033[1;32m== demo stack is UP, drive it from the UI ==\033[0m
  RPC:     $RPC   (chain id $CHAIN_ID)
  Diamond: $VAULT

Next, in another terminal:
  cd $UI_REPO && npm run dev      # http://localhost:3000

In the browser wallet (MetaMask/Rabby):
  - Add network: RPC $RPC, Chain ID $CHAIN_ID, symbol ETH
  - Import a USER key (Alice/Bob/Carol) to deposit/withdraw, OR the owner key
    $DEPLOYER_KEY  (owner+curator: Admin + Curator consoles)

Keeper is allocating/rebalancing autonomously. Ctrl-C here tears it all down.
EOF
wait
