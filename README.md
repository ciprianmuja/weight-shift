# `weight-shift`

## Abstract

Blockchain networks often leverage validators to maintain the security and integrity of the system. Typically, voting power is determined by the amount of stake a validator has or other static metrics. This documentation proposes an innovative approach to decide a validator's voting power based on their on-chain activity, ensuring that active and dedicated validators are rewarded appropriately.

## Contents
**Overview**
* [QueryData](#querydata)
    * [Monitor On-Chain Activity - PrepareProposal()](#prepareproposal)
* [SendData](#senddata)
    * [Send Data to be computed - ExtendVote()](#extendvote)
* [ComputeData](#computedata)
    * [Send Data to be computed - ExtendVote()](#extendvote)



## QueryData
### Monitor On-Chain Activity - PrepareProposal()
In the PrepareProposal step, the system will query the data related to the on-chain activity of validators. This data encompasses the number of transactions validated, proposed blocks, uptime percentage, and any other relevant metrics that contribute to the on-chain activity.
- Data will be queried from the latest block received

- Track Uptime: Monitor the operational status of each validator's node and record the uptime percentage.
- Log Proposed Blocks: Count and log the number of blocks proposed by each validator.
- Record Transactions Validated: Track and record the number of transactions validated by each validator.
- Detect Slashing Events: Identify and log any instances where validators are penalized for misbehavior.

## SendData
### Send Data to be computed - ExtendVote()
Once the data has been aggregated, the ExtendVote() function serves as the transmission mechanism. Through this function, all relevant data about a validator's on-chain activity is transmitted to be processed.
The primary objective of ExtendVote() is to extend the traditional voting mechanism by encapsulating additional data - in this case, the on-chain activity metrics. This enriched vote, carrying more context, then becomes the basis for adjusting voting power.

## ComputeData
### Compute Received data - PreBlocker()
With the data in place from the ExtendVote step, the PreBlocker() function will be invoked. This function will process the data and update the blockchain state, effectively adjusting the voting power of validators based on their on-chain activity.

- Determine Base Voting Power: Establish a minimum voting power assigned to each validator, ensuring even newly joined or less active validators have a baseline influence.
- Calculate Activity Bonus: Based on the tracked on-chain activity, compute additional bonuses to the voting power:
  - Uptime Bonus
  - Proposed Blocks Bonus
  - Transactions Validated Bonus
- Apply Slashing Penalties: Deduct penalties from the voting power of validators who have engaged in malicious or negligent actions.
- Sync with ABCI++: Integrate the voting power computations with ABCI++ to ensure that the updated voting powers are recognized and applied in consensus rounds.
