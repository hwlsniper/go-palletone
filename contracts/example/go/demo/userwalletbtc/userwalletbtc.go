package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/naoina/toml"

	"github.com/palletone/btc-adaptor"
)

type BTCConfig struct {
	NetID        int
	Host         string
	RPCUser      string
	RPCPasswd    string
	CertPath     string
	WalletPasswd string
}
type MyWallet struct {
	BtcConfig   BTCConfig
	NameKey     map[string]string
	NamePubkey  map[string]string
	NameAddress map[string]string
	AddressKey  map[string]string
}

var gWallet = NewWallet()
var gWalletFile = "./btcwallet.toml"

var gTomlConfig = toml.DefaultConfig

func NewWallet() *MyWallet {
	return &MyWallet{
		BtcConfig: BTCConfig{
			NetID:        1,
			Host:         "localhost:18332",
			RPCUser:      "zxl",
			RPCPasswd:    "123456",
			CertPath:     "",
			WalletPasswd: "1",
		},
		NameKey:     map[string]string{},
		NamePubkey:  map[string]string{},
		NameAddress: map[string]string{},
		AddressKey:  map[string]string{}}
}

func loadConfig(file string, w *MyWallet) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = gTomlConfig.NewDecoder(bufio.NewReader(f)).Decode(w)
	return err
}

func saveConfig(file string, w *MyWallet) error {
	configFile, err := os.Create(file)
	defer configFile.Close()
	if err != nil {
		return err
	}

	configToml, err := gTomlConfig.Marshal(w)
	if err != nil {
		return err
	}

	_, err = configFile.Write(configToml)
	if err != nil {
		return err
	}

	return nil
}

func createKey(name string) error {
	var btcadaptor adaptorbtc.AdaptorBTC
	btcadaptor.NetID = gWallet.BtcConfig.NetID
	//
	key := btcadaptor.NewPrivateKey()
	gWallet.NameKey[name] = key

	//
	pubkey := btcadaptor.GetPublicKey(key)
	gWallet.NamePubkey[name] = pubkey
	fmt.Println(name, "'s pubkey : ", pubkey)

	//
	addr := btcadaptor.GetAddress(key)
	gWallet.NameAddress[name] = addr
	gWallet.AddressKey[addr] = key

	return saveConfig(gWalletFile, gWallet)
}

func giveAlice(txid string, index string, amount string, fee string, prikey string) error {
	//
	vout, err := strconv.Atoi(index)
	if err != nil {
		return errors.New("Index is Invalid.")
	}

	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return errors.New("Amount is Invalid.")
	}
	feeValue, err := strconv.ParseFloat(fee, 64)
	if err != nil {
		return errors.New("Fee is Invalid.")
	}
	//
	var rawTransactionGenParams adaptorbtc.RawTransactionGenParams
	rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptorbtc.Input{txid, uint32(vout)})
	rawTransactionGenParams.Outputs = append(rawTransactionGenParams.Outputs, adaptorbtc.Output{gWallet.NameAddress["alice"], amountValue - feeValue})
	//
	var btcadaptor adaptorbtc.AdaptorBTC
	btcadaptor.NetID = gWallet.BtcConfig.NetID
	btcadaptor.Host = gWallet.BtcConfig.Host
	btcadaptor.RPCUser = gWallet.BtcConfig.RPCUser
	btcadaptor.RPCPasswd = gWallet.BtcConfig.RPCPasswd
	btcadaptor.CertPath = gWallet.BtcConfig.CertPath
	//
	rawResult, err := btcadaptor.RawTransactionGen(&rawTransactionGenParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(rawResult)
	}
	//
	var rawTransactionGenResult adaptorbtc.RawTransactionGenResult
	err = json.Unmarshal([]byte(rawResult), &rawTransactionGenResult)
	if err != nil {
		fmt.Println(err.Error())
	}
	//
	var signTxSendParams adaptorbtc.SignTxSendParams
	signTxSendParams.TransactionHex = rawTransactionGenResult.Rawtx
	signTxSendParams.Privkeys = append(signTxSendParams.Privkeys, prikey)
	sendReuslt, err := btcadaptor.SignTxSend(&signTxSendParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(sendReuslt)
	}

	return nil
}

func aliceSendBTCToMultiSigAddr(txid string, index string, amount string, fee string, multiSigAddr string) error {
	//
	vout, err := strconv.Atoi(index)
	if err != nil {
		return errors.New("Index is Invalid.")
	}

	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return errors.New("Amount is Invalid.")
	}
	feeValue, err := strconv.ParseFloat(fee, 64)
	if err != nil {
		return errors.New("Fee is Invalid.")
	}
	//
	var rawTransactionGenParams adaptorbtc.RawTransactionGenParams
	rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptorbtc.Input{txid, uint32(vout)})
	rawTransactionGenParams.Outputs = append(rawTransactionGenParams.Outputs, adaptorbtc.Output{multiSigAddr, amountValue - feeValue})
	//
	var btcadaptor adaptorbtc.AdaptorBTC
	btcadaptor.NetID = gWallet.BtcConfig.NetID
	btcadaptor.Host = gWallet.BtcConfig.Host
	btcadaptor.RPCUser = gWallet.BtcConfig.RPCUser
	btcadaptor.RPCPasswd = gWallet.BtcConfig.RPCPasswd
	btcadaptor.CertPath = gWallet.BtcConfig.CertPath
	//
	rawResult, err := btcadaptor.RawTransactionGen(&rawTransactionGenParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(rawResult)
	}
	//
	var rawTransactionGenResult adaptorbtc.RawTransactionGenResult
	err = json.Unmarshal([]byte(rawResult), &rawTransactionGenResult)
	if err != nil {
		fmt.Println(err.Error())
	}
	//
	var signTxSendParams adaptorbtc.SignTxSendParams
	signTxSendParams.TransactionHex = rawTransactionGenResult.Rawtx
	signTxSendParams.Privkeys = append(signTxSendParams.Privkeys, gWallet.NameKey["alice"])
	sendReuslt, err := btcadaptor.SignTxSend(&signTxSendParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(sendReuslt)
	}

	return nil
}

func bobSpendBTCFromMultiAddr(txid string, index string, amount string, fee string, redeem string) error {
	//
	vout, err := strconv.Atoi(index)
	if err != nil {
		return errors.New("Index is Invalid.")
	}

	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return errors.New("Amount is Invalid.")
	}
	feeValue, err := strconv.ParseFloat(fee, 64)
	if err != nil {
		return errors.New("Fee is Invalid.")
	}

	//
	var btcadaptor adaptorbtc.AdaptorBTC
	btcadaptor.NetID = gWallet.BtcConfig.NetID
	bobAddr := btcadaptor.GetAddress(gWallet.NameKey["bob"])
	//
	var rawTransactionGenParams adaptorbtc.RawTransactionGenParams
	rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptorbtc.Input{txid, uint32(vout)})
	rawTransactionGenParams.Outputs = append(rawTransactionGenParams.Outputs, adaptorbtc.Output{bobAddr, amountValue - feeValue})
	//
	btcadaptor.Host = gWallet.BtcConfig.Host
	btcadaptor.RPCUser = gWallet.BtcConfig.RPCUser
	btcadaptor.RPCPasswd = gWallet.BtcConfig.RPCPasswd
	btcadaptor.CertPath = gWallet.BtcConfig.CertPath
	//
	rawResult, err := btcadaptor.RawTransactionGen(&rawTransactionGenParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(rawResult)
	}
	//
	var rawTransactionGenResult adaptorbtc.RawTransactionGenResult
	err = json.Unmarshal([]byte(rawResult), &rawTransactionGenResult)
	if err != nil {
		fmt.Println(err.Error())
	}
	//
	var signTxParams adaptorbtc.SignTransactionParams
	signTxParams.TransactionHex = rawTransactionGenResult.Rawtx
	signTxParams.RedeemHex = redeem
	signTxParams.Privkeys = append(signTxParams.Privkeys, gWallet.NameKey["bob"])
	signReuslt, err := btcadaptor.SignTransaction(&signTxParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(signReuslt)
	}

	return nil
}
func main() {
	err := loadConfig(gWalletFile, gWallet)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	args := os.Args
	cmd := strings.ToLower(args[1])

	switch cmd {
	case "init":
		createKey("alice")
		createKey("bob")
	case "give":
		if len(args) < 7 {
			fmt.Println("Params : give, txid, index, amount, fee, prikey")
			return
		}
		err := giveAlice(args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "alice":
		if len(args) < 7 {
			fmt.Println("Params : alice, txid, index, amount, fee, multiSigAddr")
			return
		}
		err := aliceSendBTCToMultiSigAddr(args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "bob":
		if len(args) < 7 {
			fmt.Println("Params : bob, txid, index, amount, fee, redeem")
			return
		}
		err := bobSpendBTCFromMultiAddr(args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			fmt.Println(err.Error())
		}
	default:
		fmt.Println("Invalid cmd.")
	}
}
