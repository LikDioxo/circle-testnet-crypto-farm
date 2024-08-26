package infra

import (
	"errors"
	"flag"
)

var MainConfig Config

type Config struct {
	DestinationAddress       string
	WalletNumber             int
	Blockchain               string
	NativeAmountModifier     int
	BalanceCheckThresholdSec int
}

func GetConfig() (Config, error) {
	MainConfig.NativeAmountModifier = 20     // default value
	MainConfig.BalanceCheckThresholdSec = 5 // default value

	flag.StringVar(&MainConfig.DestinationAddress, "dest", "", "Destination address where to send crypto")
	flag.IntVar(&MainConfig.WalletNumber, "n", 1, "Number of intermediate wallets to create. Directry affects how much crypto will be farmed in single run")
	flag.StringVar(&MainConfig.Blockchain, "blockchain", "", "Blockchain. Examples: MATIC-AMOY, ETH-SEPOLIA")

	flag.Parse()

	if MainConfig.DestinationAddress == "" {
		return MainConfig, errors.New("destination address is mandatory paramenter and can't be an empty string")
	}

	if MainConfig.Blockchain == "" {
		return MainConfig, errors.New("blockchain is mandatory paramenter and can't be an empty string")
	}

	return MainConfig, nil
}