package app

import (
	"circle-cryto-farm/internal/infra"
	"circle-cryto-farm/internal/infra/env"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type GetPublicKeyResponse struct {
	Data struct {
		PublicKey string
	}
}

type CreateWalletSetResponse struct {
	Data struct {
		WalletSet struct {
			CreatedAt   string
			CustodyType string
			Id          string
			Name        string
			UpdateDate  string
		}
	}
}

type CreateWalletsResponse struct {
	Data struct {
		Wallets []Wallet
	}
}

type Wallet struct {
	Id          string
	State       string
	WalletSetId string
	CustodyType string
	Address     string
	Blockchain  string
	AccountType string
	UpdateDate  string
	CreateDate  string
}

type GetWalletBalanceResponse struct {
	Data struct {
		TokenBalances []TokenBalance
	}
}

type TokenBalance struct {
	Token struct {
		Id       string
		Symbol   string
		IsNative bool
	}
	Amount string
}

func FetchPublicKey() (*GetPublicKeyResponse, error) {
	url := os.Getenv(env.CIRCLE_API_URL) + "/w3s/config/entity/publicKey"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv(env.API_KEY)))

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return nil, errors.New(string(body))
	}

	var decoded GetPublicKeyResponse
	err := json.Unmarshal(body, &decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key: %w", err)
	}

	return &decoded, nil
}

func ParseRsaPublicKeyFromPem(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
	}
	return nil, errors.New("key type is not rsa")
}

func EncryptOAEP(pubKey *rsa.PublicKey, message []byte) (ciphertext []byte, err error) {
	random := rand.Reader
	ciphertext, err = rsa.EncryptOAEP(sha256.New(), random, pubKey, message, nil)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

func GenerateCiphertext() (string, error) {
	entitySecret, err := hex.DecodeString(os.Getenv(env.ENTITY_SECRET))
	if err != nil {
		return "", err
	}
	if len(entitySecret) != 32 {
		panic("invalid entity secret")
	}
	pubKey, err := ParseRsaPublicKeyFromPem([]byte(os.Getenv(env.PUBLIC_KEY)))
	if err != nil {
		return "", err
	}
	cipher, err := EncryptOAEP(pubKey, entitySecret)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cipher), nil
}

func CreateWalletSet() (*CreateWalletSetResponse, error) {
	idempotencyKey := uuid.NewString()
	name := idempotencyKey[:8]
	url := os.Getenv(env.CIRCLE_API_URL) + "/w3s/developer/walletSets"
	cipherText, err := GenerateCiphertext()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CipherText: %w", err)
	}

	payload := strings.NewReader(fmt.Sprintf(
		`{"idempotencyKey":"%s", "entitySecretCipherText":"%s", "name":"%s"}`,
		idempotencyKey,
		cipherText,
		name,
	))

	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv(env.API_KEY)))
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != 201 {
		return nil, errors.New(string(body))
	}

	var decoded CreateWalletSetResponse
	err = json.Unmarshal(body, &decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet set: %w", err)
	}

	return &decoded, nil
}

func CreateWallets(walletSetId string, n int, blockchain string) (*CreateWalletsResponse, error) {
	idempotencyKey := uuid.NewString()
	url := os.Getenv(env.CIRCLE_API_URL) + "/w3s/developer/wallets"
	cipherText, err := GenerateCiphertext()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CipherText: %w", err)
	}

	payload := strings.NewReader(fmt.Sprintf(
		`{"idempotencyKey":"%s","entitySecretCipherText":"%s","blockchains":["%s"],"count":%d,"walletSetId":"%s"}`,
		idempotencyKey,
		cipherText,
		blockchain,
		n,
		walletSetId,
	))

	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv(env.API_KEY)))

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var decoded CreateWalletsResponse

	err = json.Unmarshal(body, &decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallets: %w", err)
	}

	return &decoded, err
}

func FundAddress(address string, blockchain string) (bool, error) {
	url := os.Getenv(env.CIRCLE_API_URL) + "/faucet/drips"

	payload := strings.NewReader(fmt.Sprintf(
		`{"address":"%s", "blockchain":"%s", "native":true, "usdc":true, "eurc": true}`,
		address,
		blockchain,
	))
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv(env.API_KEY)))

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != 204 {
		return false, errors.New(string(body))
	}

	return true, nil
}

func GetWalletBalance(id string) (*GetWalletBalanceResponse, error) {
	url := fmt.Sprintf("%s/w3s/wallets/%s/balances", os.Getenv(env.CIRCLE_API_URL), id)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv(env.API_KEY)))

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return nil, errors.New(string(body))
	}

	var decoded GetWalletBalanceResponse

	err := json.Unmarshal(body, &decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	return &decoded, nil
}

func MakeTransaction(
	walletId string,
	tokenId string,
	amount string,
	destination string,
) (bool, error) {
	idempotencyKey := uuid.NewString()
	url := os.Getenv(env.CIRCLE_API_URL) + "/w3s/developer/transactions/transfer"
	cipherText, err := GenerateCiphertext()
	if err != nil {
		return false, fmt.Errorf("failed to generate CipherText: %w", err)
	}

	payload := strings.NewReader(fmt.Sprintf(
		`{"idempotencyKey":"%s","entitySecretCipherText":"%s","amounts":["%s"],"feeLevel":"MEDIUM","tokenId":"%s","walletId":"%s","destinationAddress":"%s"}`,
		idempotencyKey,
		cipherText,
		amount,
		tokenId,
		walletId,
		destination,
	))

	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv(env.API_KEY)))

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != 201 {
		return false, errors.New(string(body))
	}

	return true, nil
}

func WaitForBalances(wallet Wallet) []TokenBalance {
	for {
		walletBalances, err := GetWalletBalance(wallet.Id)
		if err != nil {
			panic(err)
		}

		if len(walletBalances.Data.TokenBalances) != 0 {
			return walletBalances.Data.TokenBalances
		}

		fmt.Printf("Waiting for %d seconds for balance to update\n", infra.MainConfig.BalanceCheckThresholdSec)

		dur := time.Duration(infra.MainConfig.NativeAmountModifier * int(time.Second))
		time.Sleep(dur)
	}
}
