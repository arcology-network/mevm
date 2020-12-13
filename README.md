# mevm
MEVM is a multi-instance enabled version of original the Ethereum EVM for Arcology network. It is fully compatible with original EVM with some enhanced features added. Solidity code running in each MEVM instance is still single-threaded. Arcology allows multiple instances of MEVM running in isolation to process multiple transactions concurrently without causing state inconsistency. 
