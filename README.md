# mevm
MEVM is a multi-instance enabled version of original the Ethereum EVM for Arcology network.  Solidity code running in each MEVM instance is still single-threaded. It is fully compatible with original EVM with some enhanced features added. Arcology allows multiple instances of MEVM running in isolation to process multiple transactions concurrently without causing state inconsistency. 
