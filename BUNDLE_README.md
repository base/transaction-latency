# Bundle Sending Implementation

This implementation adds support for sending transaction bundles using the `eth_sendBundle` RPC method, compatible with the Base TIPS (Transaction Inclusion Protocol for Sequencers) format.

## Features Added

### 1. Bundle Data Structure
```go
type Bundle struct {
    Txs                   []string  `json:"txs"`                             // Raw transaction bytes (hex-encoded)
    BlockNumber           uint64    `json:"blockNumber"`                     // Target block number
    FlashblockNumberMin   *uint64   `json:"flashblockNumberMin,omitempty"`   // Optional: minimum flashblock number
    FlashblockNumberMax   *uint64   `json:"flashblockNumberMax,omitempty"`   // Optional: maximum flashblock number
    MinTimestamp          *uint64   `json:"minTimestamp,omitempty"`          // Optional: minimum timestamp
    MaxTimestamp          *uint64   `json:"maxTimestamp,omitempty"`          // Optional: maximum timestamp
    RevertingTxHashes     []string  `json:"revertingTxHashes"`               // Transaction hashes that can revert
    ReplacementUuid       *string   `json:"replacementUuid,omitempty"`       // Optional: replacement UUID
    DroppingTxHashes      []string  `json:"droppingTxHashes"`                // Transaction hashes to drop
}
```

### 2. Core Functions

#### `sendBundle(client, signedTxs, targetBlockNumber)`
- Converts signed transactions to hex-encoded raw transaction data
- Creates Bundle structure matching Base TIPS format
- Sends bundle via `eth_sendBundle` RPC call
- Returns bundle hash on success

#### `createAndSendBundle(chainId, privateKey, fromAddress, toAddress, client, numTxs)`
- Creates multiple transactions with sequential nonces
- Signs all transactions
- Targets the next block
- Calls `sendBundle` to submit the bundle

## Environment Variables

Add these to your `.env` file:

```bash
# Enable bundle testing
RUN_BUNDLE_TEST=true

# Number of transactions per bundle (default: 3)
BUNDLE_SIZE=5

# Other existing variables remain the same
PRIVATE_KEY=your_private_key_here
TO_ADDRESS=0x...
FLASHBLOCKS_URL=https://sepolia-preconf.base.org
BASE_URL=https://sepolia.base.org
REGION=texas
```

## Usage

### 1. Enable Bundle Testing
Set `RUN_BUNDLE_TEST=true` in your `.env` file.

### 2. Configure Bundle Size
Set `BUNDLE_SIZE=N` where N is the number of transactions you want in each bundle.

### 3. Run the Application
```bash
go run main.go
```

### 4. Expected Output
```
Starting bundle test with 3 transactions per bundle
Created transaction 0 with nonce 42, hash: 0x...
Created transaction 1 with nonce 43, hash: 0x...
Created transaction 2 with nonce 44, hash: 0x...
Bundle sent successfully with hash: 0x...
Bundle sent with hash: 0x..., targeting block: 12345
Bundle test completed successfully
```

## Technical Details

### Bundle Atomicity
- All transactions in a bundle are either included together or not at all
- If one transaction fails, the entire bundle is rejected
- Bundles target specific block numbers

### Nonce Management
- Transactions use sequential nonces starting from the current confirmed nonce
- Each transaction in the bundle increments the nonce by 1

### Gas Settings
- Uses the same gas calculation logic as individual transactions
- Applies 20% buffer to gas tip
- Sets gas fee cap to 2x the suggested gas price

### Block Targeting
- Automatically targets the next block (`currentBlock + 1`)
- MEV searchers/builders will attempt inclusion if profitable

## Compatibility

This implementation is compatible with:
- Base TIPS (Transaction Inclusion Protocol for Sequencers)
- Base Sepolia testnet with preconf
- Any RPC endpoint that supports `eth_sendBundle`

## Error Handling

The implementation includes comprehensive error handling for:
- Transaction signing failures
- RPC call failures
- Network connectivity issues
- Invalid bundle parameters

## Testing

To test bundle functionality:

1. Ensure you have sufficient ETH in your account for gas fees
2. Set `RUN_BUNDLE_TEST=true`
3. Start with a small `BUNDLE_SIZE` (e.g., 2-3 transactions)
4. Monitor logs for bundle submission and confirmation