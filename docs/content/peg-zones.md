# Pegzones

## Components and Roles

// TODO

pegzone

Signers: monitoring the peg zone then sign those IBC transactions, effectively converting the signature scheme to Ethereum-understandable private keys, in secp256k1 format. Your transaction has just been signed on the peg zone.

Relayers: watch the peg zone and wait until they see that +2/3 signers have signed the transaction before batching the signed transaction into a list of all the other transactions sent through IBC. They then relay the signature-appended list to the EVM where the Ethereum smart contract lives.

Witness: they watch logs for events coming from the Ethereum smart contract
