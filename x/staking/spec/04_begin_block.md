<!--
order: 4
-->

# Begin-Block

Each abci begin block call, the historical info will get stored and pruned
according to the `HistoricalEntries` parameter. This calls also decides whether
end block will process staking transactions, submitted over the last aeon period,
for computing validator set updates. Two separate updates, for the DKG and consensus 
validators, are passed down to Tendermint at the boundary between aeons.

## Historical Info Tracking

If the `HistoricalEntries` parameter is 0, then the `BeginBlock` performs a no-op.

Otherwise, the latest historical info is stored under the key `historicalInfoKey|height`, while any entries older than `height - HistoricalEntries` is deleted.
In most cases, this results in a single entry being pruned per block.
However, if the parameter `HistoricalEntries` has changed to a lower value there will be multiple entries in the store that must be pruned.

## Triggering Validator Set Changes

If the block height of the current block is equal to `NextAeonStart - 1` then DKG validator updates are triggered in the corresponding `EndBlock`.  For height equal to `NextAeonStart - 2` the consensus validators are updated. Delaying validator set updates to aeon boundaries can be turned off by setting `delayValidatorUpdates` to false in the staking keeper. Validator updates are computed on every `EndBlock`.
