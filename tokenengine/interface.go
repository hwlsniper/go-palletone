package tokenengine

import (
	"github.com/palletone/go-palletone/common"

	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/core/accounts/keystore"
	"go/types"
)

type TokenEngine interface {
	//Lock script
	GenerateP2PKHLockScript(pubKeyHash []byte) []byte
	GenerateP2SHLockScript(redeemScriptHash []byte) []byte
	GenerateRedeemScript(needed byte, pubKeys [][]byte) []byte
	GenerateLockScript(address common.Address) []byte

	//Unlock script
	GenerateP2PKHUnlockScript(sign []byte, pubKey []byte) []byte
	GenerateP2SHUnlockScript(signs [][]byte, redeemScript []byte) []byte
	ScriptValidate(utxoLockScript []byte, utxoAmount int64, tx *modules.PaymentPayload, inputIndex int) error
	ParseAddress(lockScript []byte) common.Address
	SignTxInput(tx *modules.Transaction,utxoLockScript []byte,inputIdex int, store keystore.KeyStore) types.Signature
}
