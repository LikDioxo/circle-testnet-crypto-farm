package main

import (
	"circle-cryto-farm/internal/app"
	"circle-cryto-farm/internal/infra"
	"fmt"
	"strconv"
)

func main() {
	config, err := infra.GetConfig()
	err = infra.LoadEnv()

	if err != nil {
		fmt.Println(err)
		return
	}

	walletSet, err := app.CreateWalletSet()
	if err != nil {
		panic(err)
	}

	fmt.Printf("WalletSet id: %s\n", walletSet.Data.WalletSet.Id)
	wallets, err := app.CreateWallets(walletSet.Data.WalletSet.Id, config.WalletNumber, config.Blockchain)
	if err != nil {
		panic(err)
	}

	for _, wallet := range wallets.Data.Wallets {
		fmt.Printf("Funding wallet: %s\n", wallet.Id)

		_, err := app.FundAddress(wallet.Address, wallet.Blockchain)
		if err != nil {
			panic(err)
		}

		tokenBalances := app.WaitForBalances(wallet)
		// Reverse because first one is native currency. I know it's lazy coding but it's 5 o'clock
		for i := len(tokenBalances) - 1; i >= 0; i-- {
			var amountToSend string

			if tokenBalances[i].Token.IsNative {
				floatNativeBalance, _ := strconv.ParseFloat(tokenBalances[i].Amount, 32)
				amountToSend = fmt.Sprintf("%f", floatNativeBalance - (floatNativeBalance*float64(config.NativeAmountModifier))/100)
			} else {
				amountToSend = tokenBalances[i].Amount
			}

			res, err := app.MakeTransaction(
				wallet.Id,
				tokenBalances[i].Token.Id,
				amountToSend,
				config.DestinationAddress,
			)

			if err != nil {
				fmt.Println(2)
				panic(err)
			}

			fmt.Printf(
				"Sending %s : %s to %s. Success: %v\n",
				amountToSend,
				tokenBalances[i].Token.Symbol,
				config.DestinationAddress,
				res,
			)
		}
	}
}
