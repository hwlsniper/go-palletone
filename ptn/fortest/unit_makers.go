/* This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.

   @author PalletOne core developers <dev@pallet.one>
   @date 2018
*/

// unit package, unit structure and storage api
package fortest

import (
	"strings"
	"time"
	"unsafe"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/storage"
)
const (
	ID_LENGTH = 32
)
var (
	PTNCOIN = IDType16{'p', 't', 'n', 'c', 'o', 'i', 'n'}
)
type IDType16 [ID_LENGTH]byte

type Units []*Unit

// key: unit.UnitHash(unit)
type Unit struct {
	UnitHeader *Header
	Txs        Transactions
	ReceivedAt   time.Time
	UnitSize   common.StorageSize
	UnitHash   common.Hash
}

func NewUnit(header *Header, txs Transactions) *Unit {
	u := &Unit{
		UnitHeader: header,
		Txs:        txs,
	}
	u.ReceivedAt = time.Now()
	u.UnitSize = u.Size()
	u.UnitHash = u.Hash()
	return u
}
func (u *Unit) Size() common.StorageSize {
	emptyUnit := Unit{}
	emptyUnit.UnitHeader = u.UnitHeader
	emptyUnit.Txs = u.Txs
	emptyUnit.ReceivedAt = u.ReceivedAt
	b, err := rlp.EncodeToBytes(emptyUnit)
	if err != nil {
		return common.StorageSize(0)
	} else {
		return common.StorageSize(len(b))
	}
}
func (u *Unit) Hash() common.Hash {
	return rlp.RlpHash(u.UnitHeader)
}
type Header struct {
	ParentsHash  []common.Hash
	AssetIDs     []IDType16
	Authors      *Authentifier
	Witness      []*Authentifier
	TxRoot       common.Hash
	Number       ChainIndex
	Extra        []byte
	Creationdate int64
}
func NewHeader(parents []common.Hash, asset []IDType16,extra []byte) *Header {
	hashs := make([]common.Hash, 0)
	hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	var b []byte
	return &Header{ParentsHash: hashs, AssetIDs: asset, Extra: append(b, extra...),Creationdate:time.Now().Unix()}
}
type Transaction struct {
	TxHash     common.Hash `json:"txhash"`
	TxMessages []Message   `json:"messages"`
	Locktime   uint32      `json:"lock_time"`
}
type Transactions []*Transaction
func NewCoinbaseTransaction() (*Transaction,error){
	input := Input{}
	output := Output{}
	payload := PaymentPayload{
		Inputs:  []Input{input},
		Outputs: []Output{output},
	}
	msg := Message{
		App:     APP_PAYMENT,
		Payload: payload,
	}
	coinbase := &Transaction{
		TxMessages: []Message{msg},
	}
	coinbase.TxHash = coinbase.Hash()
	return coinbase,nil
}
func (tx *Transaction) Hash() common.Hash {
	if tx.TxHash != (common.Hash{}) {
		return tx.TxHash
	}
	return rlp.RlpHash(tx)
}

//


func newDag(memdb ptndb.Database,gunit *Unit,number int) (Units,error){
	units := make(Units,number)
	par := gunit
	fmt.Println("len(units).........",len(units))
	fmt.Println("number.........",number)
	for i := 0; i < number;i++ {
		fmt.Println("createUnit",i)
		header := NewHeader([]common.Hash{par.UnitHash},[]IDType16{PTNCOIN},[]byte{})
		header.Number.AssetID = par.UnitHeader.Number.AssetID
		header.Number.IsMain = par.UnitHeader.Number.IsMain
		header.Number.Index = par.UnitHeader.Number.Index+1
		header.Authors = &Authentifier{"P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ",[]byte{},[]byte{},[]byte{}}
		header.Witness = []*Authentifier{&Authentifier{"P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ",[]byte{},[]byte{},[]byte{}}}
		tx,_:= NewCoinbaseTransaction()
		txs := Transactions{tx}
		unit := NewUnit(header,txs)
		//fmt.Println("start saveUnit")
		err := SaveUnit(memdb,unit)
		if err != nil {
			fmt.Println("Save==",err)
		}
		fmt.Printf("--------这是父块----unit-----%#v\n",unit)
		fmt.Printf("--------这是父块----unit.UnitHeader-----%#v\n",unit.UnitHeader)
		fmt.Printf("--------这是父块----unit.Txs-----%#v\n", unit.Txs[0].Hash())
		fmt.Printf("--------这是父块----unit.UnitHash-----%#v\n",unit.UnitHash)
		fmt.Printf("--------这是父块----unit.UnitHeader.ParentsHash-----%#v\n",unit.UnitHeader.ParentsHash)
		fmt.Printf("--------这是父块----unit.UnitHeader.Number.Index-----%#v\n",unit.UnitHeader.Number.Index)
		fmt.Println("createUnit",i)
		units[i] = unit
		par = unit
	}
	return units,nil
}
/**
保存单元数据，如果单元的结构基本相同
save genesis unit data
*/
func SaveUnit(db ptndb.Database,unit *Unit) error {
	if unit.UnitSize == 0 || unit.Size() == 0 {
		log.Error("Unit is null")
		return fmt.Errorf("Unit is null")
	}
	if unit.UnitSize != unit.Size() {
		log.Info("Validate size", "error", "Size is invalid")
		return fmt.Errorf("Unit size error")
	}
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	//fmt.Println("= start==SaveHeader(db,unit.UnitHash, unit.UnitHeader)======")
	//fmt.Printf("--------这是最新块----unit-----%#v\n", unit.UnitHash)
	//fmt.Printf("--------这是最新块----unit.UnitHeader-----%#v\n", unit.UnitHeader)
	if err := SaveHeader(db,unit.UnitHash, unit.UnitHeader); err != nil {
		log.Info("SaveHeader:", "error", err.Error())
		//fmt.Println("这是最新块",err.Error())
		return fmt.Errorf("Unit header save error")
	}
	//fmt.Println("= end==SaveHeader(db,unit.UnitHash, unit.UnitHeader)======")
	// key is like "[UNIT_HASH_NUMBER][unit_hash]"
	//fmt.Printf("unit.UnitHeader.Number==%#v",unit.UnitHeader.Number)
	if err := SaveHashNumber(db,unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Info("SaveHashNumber:", "error", err.Error())
		return fmt.Errorf("Save unit hash and number error")
	}
	if err := SaveTransaction(db,unit.UnitHash,unit.Txs[0]); err != nil {
		log.Info("Save transaction:", "error", err.Error())
		return err
	}
	storage.PutCanonicalHash(db, unit.UnitHash, unit.NumberU64())
	storage.PutHeadHeaderHash(db, unit.UnitHash)
	storage.PutHeadUnitHash(db, unit.UnitHash)
	storage.PutHeadFastUnitHash(db, unit.UnitHash)
	return nil
}


func  SaveGenesis(db ptndb.Database,unit *Unit) error{
	//fmt.Println("unit.NumberU64()====",unit.NumberU64())
	if unit.NumberU64() != 0 {
		return fmt.Errorf("can't commit genesis unit with number > 0")
	}
	//fmt.Println("start saveUnit")
	err := SaveUnit(db,unit)
	if err != nil {
		fmt.Println("SaveGenesis==",err)
	}
	//fmt.Println("end saveUnit")
	return  nil
}
func (cpy *Header) CopyHeader(h *Header) {
	cpy = h
	if len(h.ParentsHash) > 0 {
		cpy.ParentsHash = make([]common.Hash, len(h.ParentsHash))
		for i := 0; i < len(h.ParentsHash); i++ {
			cpy.ParentsHash[i] = h.ParentsHash[i]
		}
	}

	if len(h.AssetIDs) > 0 {
		cpy.AssetIDs = make([]IDType16, len(h.AssetIDs))
		for i := 0; i < len(h.AssetIDs); i++ {
			cpy.AssetIDs[i] = h.AssetIDs[i]
		}
	}

}



func (h *Header) Index() uint64 {
	return h.Number.Index
}
func (h *Header) ChainIndex() ChainIndex {
	return h.Number
}



func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)/8)
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
//func CopyHeader(h *Header) *Header {
//	cpy := *h
//
//	if len(h.ParentsHash) > 0 {
//		cpy.ParentsHash = make([]common.Hash, len(h.ParentsHash))
//		for i := 0; i < len(h.ParentsHash); i++ {
//			cpy.ParentsHash[i].Set(h.ParentsHash[i])
//		}
//	}
//
//	if len(h.AssetIDs) > 0 {
//		copy(cpy.AssetIDs, h.AssetIDs)
//	}
//
//	if len(h.Witness) > 0 {
//		copy(cpy.Witness, h.Witness)
//	}
//
//	if len(h.TxRoot) > 0 {
//		cpy.TxRoot.Set(h.TxRoot)
//	}
//
//	return &cpy
//}

//func (u *Unit) CopyBody(txs Transactions) Transactions {
//	if len(txs) > 0 {
//		u.Txs = make([]*Transaction, len(txs))
//		for i, pTx := range txs {
//			tx := Transaction{
//				TxHash:   pTx.TxHash,
//				Locktime: pTx.Locktime,
//			}
//			if len(pTx.TxMessages) > 0 {
//				tx.TxMessages = make([]Message, len(pTx.TxMessages))
//				for j := 0; j < len(pTx.TxMessages); j++ {
//					tx.TxMessages[j] = pTx.TxMessages[j]
//				}
//			}
//			u.Txs[i] = &tx
//		}
//	}
//	return u.Txs
//}

//wangjiyou add for ptn/fetcher.go


type OutPoint struct {
	TxHash       common.Hash // reference Utxo struct key field
	MessageIndex uint32      // message index in transaction
	OutIndex     uint32
}

func (unit *Unit) IsEmpty() bool {
	if unit == nil || len(unit.Txs) <= 0 {
		return true
	}
	return false
}
type TxPoolTransaction struct {
	Tx *Transaction

	CreationDate string  `json:"creation_date" rlp:"-"`
	Priority_lvl float64 `json:"priority_lvl" rlp:"-"` // 打包的优先级
	Nonce        uint64  // transaction'hash maybe repeat.
	Pending      bool
	Confirmed    bool
	Extra        []byte
}

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Transactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}
type TxPoolTxs []*TxPoolTransaction


type ChainIndex struct {
	AssetID IDType16
	IsMain  bool
	Index   uint64
}

func (height ChainIndex) String() string {
	data, err := rlp.EncodeToBytes(height)
	if err != nil {
		return ""
	}
	return string(data)
}
func (height ChainIndex) Bytes() []byte {
	data, err := rlp.EncodeToBytes(height)
	if err != nil {
		return nil
	}
	return data[:]
}

const (
	APP_PAYMENT         = 0x01
	APP_CONTRACT_TPL    = 0x02
	APP_CONTRACT_DEPLOY = 0x03
	APP_CONTRACT_INVOKE = 0x04
	APP_CONFIG          = 0x05
	APP_TEXT            = 0x06
)

// key: message.UnitHash(message+timestamp)
type Message struct {
	App     byte        `json:"app"`     // message type
	Payload interface{} `json:"payload"` // the true transaction data
}

// return message struct
func NewMessage(app byte, payload interface{}) *Message {
	m := new(Message)
	m.App = app
	m.Payload = payload
	return m
}
//
//func (msg *Message) CopyMessages(cpyMsg *Message) *Message {
//	msg.App = cpyMsg.App
//	msg.Payload = cpyMsg.Payload
//	switch cpyMsg.App {
//	case APP_PAYMENT, APP_CONTRACT_TPL, APP_TEXT:
//		msg.Payload = cpyMsg.Payload
//	case APP_CONFIG:
//		payload, _ := cpyMsg.Payload.(ConfigPayload)
//		newPayload := ConfigPayload{
//			ConfigSet: []PayloadMapStruct{},
//		}
//		for _, p := range payload.ConfigSet {
//			newPayload.ConfigSet = append(newPayload.ConfigSet, PayloadMapStruct{Key: p.Key, Value: p.Value})
//		}
//		msg.Payload = newPayload
//	case APP_CONTRACT_DEPLOY:
//		payload, _ := cpyMsg.Payload.(ContractDeployPayload)
//		newPayload := ContractDeployPayload{
//			TemplateId:   payload.TemplateId,
//			ContractId:   payload.ContractId,
//			Args:         payload.Args,
//			Excutiontime: payload.Excutiontime,
//		}
//		readSet := []ContractReadSet{}
//		for _, rs := range payload.ReadSet {
//			readSet = append(readSet, ContractReadSet{Key: rs.Key, Value: &StateVersion{Height: rs.Value.Height, TxIndex: rs.Value.TxIndex}})
//		}
//		writeSet := []PayloadMapStruct{}
//		for _, ws := range payload.WriteSet {
//			writeSet = append(writeSet, PayloadMapStruct{Key: ws.Key, Value: ws.Value})
//		}
//		newPayload.ReadSet = readSet
//		newPayload.WriteSet = writeSet
//		msg.Payload = newPayload
//	case APP_CONTRACT_INVOKE:
//		payload, _ := cpyMsg.Payload.(ContractInvokePayload)
//		newPayload := ContractInvokePayload{
//			ContractId:   payload.ContractId,
//			Args:         payload.Args,
//			Excutiontime: payload.Excutiontime,
//		}
//		readSet := []ContractReadSet{}
//		for _, rs := range payload.ReadSet {
//			readSet = append(readSet, ContractReadSet{Key: rs.Key, Value: &StateVersion{Height: rs.Value.Height, TxIndex: rs.Value.TxIndex}})
//		}
//		writeSet := []PayloadMapStruct{}
//		for _, ws := range payload.WriteSet {
//			writeSet = append(writeSet, PayloadMapStruct{Key: ws.Key, Value: ws.Value})
//		}
//		newPayload.ReadSet = readSet
//		newPayload.WriteSet = writeSet
//		msg.Payload = newPayload
//	}
//	return msg
//}

/************************** Payload Details ******************************************/
type PayloadMapStruct struct {
	IsDelete bool
	Key      string
	Value    interface{}
}
type Asset struct {
	AssertId IDType16 `json:"assert_id"` // 资产类别
	UniqueId IDType16 `json:"unique_id"` // every token has its unique id
	ChainId  uint64   `json:"chain_id"`  // main chain id or sub-chain id
}
type Output struct {
	Value    uint64
	PkScript []byte
	Asset    Asset
}
type Input struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
	Extra            []byte // if user creating a new asset, this field should be it's config data. Otherwise it is null.
}
// Token exchange message and verify message
// App: payment
type PaymentPayload struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}
func (it *IDType16) SetBytes(b []byte) {
	if len(b) > len(it) {
		b = b[len(b)-ID_LENGTH:]
	}

	copy(it[ID_LENGTH-len(b):], b)
}
/**
从RLP的解码中解析出对应的payload
//*/
//func (pl *PaymentPayload) ExtractFrInterface(data interface{}) error {
//	// check data
//	fields, ok := data.([]interface{})
//	if !ok {
//		return fmt.Errorf("Data error, should be []interface{}")
//	}
//	if len(fields) != 2 {
//		return fmt.Errorf("Data is not type of PaymentPayload")
//	}
//	// extract inputs
//	txins, ok := fields[0].([]interface{})
//	if !ok {
//		return fmt.Errorf("Data is not type of PaymentPayload: invalid inputs")
//	}
//	fmt.Println("txins:", txins)
//	pl.Inputs = []Input{}
//
//	for _, in := range txins {
//		// extract one input
//		input, ok := in.([]interface{})
//		if !ok || len(input) != 3 {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid input")
//		}
//		outpoint, ok := input[0].([]interface{})
//		if !ok || len(outpoint) != 3 {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid outpoint")
//		}
//		// extract outpoint
//		txHash := common.Hash{}
//		txHash.SetBytes(outpoint[0].([]byte))
//		if _, ok := outpoint[1].([]byte); !ok {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid outpoint -1")
//		}
//		if _, ok := outpoint[2].([]byte); !ok {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid outpoint -2")
//		}
//		// extract output message index
//		msgIndex := binary.BigEndian.Uint32(FillBytes(outpoint[1].([]byte), 4))
//		// extract output out index
//		outIndex := binary.BigEndian.Uint32(FillBytes(outpoint[2].([]byte), 4))
//		// extract signature
//		sig, ok := input[1].([]byte)
//		if !ok {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid signature")
//		}
//		// extract extra data
//		extra, ok := input[2].([]byte)
//		if !ok {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid extra")
//		}
//		// save input
//		newInput := Input{
//			PreviousOutPoint: OutPoint{
//				TxHash:       txHash,
//				MessageIndex: msgIndex,
//				OutIndex:     outIndex,
//			},
//			SignatureScript: sig,
//			Extra:           extra,
//		}
//		pl.Inputs = append(pl.Inputs, newInput)
//	}
//	// extract outputs
//	txouts, ok := fields[1].([]interface{})
//	if !ok {
//		return fmt.Errorf("Data is not type of PaymentPayload: invalid outputs")
//	}
//	pl.Outputs = []Output{}
//	for _, out := range txouts {
//		// extract one output
//		output, ok := out.([]interface{})
//		if !ok || len(output) != 3 {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid output")
//		}
//		// extract output value
//		if _, ok := output[0].([]byte); !ok {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid output value")
//		}
//		val := binary.BigEndian.Uint64(FillBytes(output[0].([]byte), 8))
//		// extract output PKScript
//		pkscript, ok := output[1].([]byte)
//		if !ok {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid output script")
//		}
//		// extract output Asset
//		asset, ok := output[2].([]interface{})
//		if !ok || len(asset) != 3 {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid output script")
//		}
//		// extract asset id
//		aid, ok := asset[0].([]byte)
//		if !ok {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid output asset id")
//		}
//		newAid := IDType16{}
//		newAid.SetBytes(aid)
//		// extract asset unique id
//		uqid, ok := asset[1].([]byte)
//		if !ok {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid output unique id")
//		}
//		newUniqueID := IDType16{}
//		newUniqueID.SetBytes(uqid)
//		// extract asset chainid id
//		if _, ok := asset[2].([]byte); !ok {
//			return fmt.Errorf("Data is not type of PaymentPayload: invalid output chain id")
//		}
//		chainId := binary.BigEndian.Uint64(FillBytes(asset[2].([]byte), 8))
//
//		newOutput := Output{
//			Value:    val,
//			PkScript: pkscript,
//			Asset: Asset{
//				AssertId: newAid,
//				UniqueId: newUniqueID,
//				ChainId:  chainId,
//			},
//		}
//		pl.Outputs = append(pl.Outputs, newOutput)
//	}
//	return nil
//}

//func NewOutPoint(hash *common.Hash, messageindex uint32, outindex uint32) *OutPoint {
//	return &OutPoint{
//		TxHash:       *hash,
//		MessageIndex: messageindex,
//		OutIndex:     outindex,
//	}
//}

// NewTxOut returns a new bitcoin transaction output with the provided
// transaction value and public key script.
//func NewTxOut(value uint64, pkScript []byte, asset Asset) *Output {
//	return &Output{
//		Value:    value,
//		PkScript: pkScript,
//		Asset:    asset,
//	}
//}
//
//type StateVersion struct {
//	Height  ChainIndex
//	TxIndex uint32
//}
//
//func (version *StateVersion) String() string {
//	data, err := rlp.EncodeToBytes(*version)
//	if err != nil {
//		return ""
//	}
//	return string(data)
//}
//
//func (version *StateVersion) ParseStringKey(key string) bool {
//	ss := strings.Split(key, "^*^")
//	if len(ss) != 3 {
//		return false
//	}
//	var v StateVersion
//	if err := rlp.DecodeBytes([]byte(ss[2]), &v); err != nil {
//		log.Error("State version parse string key", "error", err.Error())
//		return false
//	}
//	version = &v
//	return true
//}

// Contract template deploy message
// App: contract_template
//type ContractTplPayload struct {
//	TemplateId []byte `json:"template_id"` // contract template id
//	Name       string `json:"name"`        // contract template name
//	Path       string `json:"path"`        // contract template execute path
//	Version    string `json:"version"`     // contract template version
//	Memery     uint16 `json:"memory"`      // coontract template bytecode memory size(Byte), use to compute transaction fee
//	Bytecode   []byte `json:"bytecode"`    // contract bytecode
//}
//
//func (tplpayload *ContractTplPayload) ExtractFrInterface(data interface{}) error {
//	// check data
//	fields, ok := data.([]interface{})
//	if !ok {
//		return fmt.Errorf("ContractTplPayload extract Data error, should be []interface{}")
//	}
//
//	if len(fields) != 6 {
//		return fmt.Errorf("ContractTplPayload extract: Data is not type of PaymentPayload")
//	}
//
//	// extract templateid
//	tplID, ok := fields[0].([]byte)
//	if !ok {
//		return fmt.Errorf("ContractTplPayload extract: invalid template id")
//	}
//	// extract name
//	name, ok := fields[1].([]byte)
//	if !ok {
//		return fmt.Errorf("ContractTplPayload extract: invalid name")
//	}
//	// extract path
//	path, ok := fields[2].([]byte)
//	if !ok {
//		return fmt.Errorf("ContractTplPayload extract: invalid path")
//	}
//	// extract version
//	version, ok := fields[3].([]byte)
//	if !ok {
//		return fmt.Errorf("ContractTplPayload extract: invalid version")
//	}
//	// extract memory
//	mem, ok := fields[4].([]byte)
//	if !ok {
//		return fmt.Errorf("ContractTplPayload extract: invalid memory")
//	}
//	memory := binary.BigEndian.Uint16(FillBytes(mem, 2))
//	// extract bytecode
//	bytecode, ok := fields[5].([]byte)
//	if !ok {
//		return fmt.Errorf("ContractTplPayload extract: invalid bytecode")
//	}
//	tplpayload.TemplateId = tplID
//	tplpayload.Name = string(name)
//	tplpayload.Path = string(path)
//	tplpayload.Version = string(version)
//	tplpayload.Memery = memory
//	tplpayload.Bytecode = bytecode
//	return nil
//}
//
//type DelContractState struct {
//	IsDelete bool
//}
//
//type ContractReadSet struct {
//	Key   string
//	Value *StateVersion
//}
//
//// Contract instance message
//// App: contract_deploy
//
//type ContractDeployPayload struct {
//	TemplateId   []byte             `json:"template_id"`   // contract template id
//	ContractId   []byte             `json:"contract_id"`   // contract id
//	Name         string             `json:"name"`          // the name for contract
//	Args         [][]byte           `json:"args"`          // contract arguments list
//	Excutiontime time.Duration      `json:"excution_time"` // contract execution time, millisecond
//	Jury         []common.Address   `json:"jury"`          // contract jurors list
//	ReadSet      []ContractReadSet  `json:"read_set"`      // the set data of read, and value could be any type
//	WriteSet     []PayloadMapStruct `json:"write_set"`     // the set data of write, and value could be any type
//}
//
//// Contract invoke message
//// App: contract_invoke
//type ContractInvokePayload struct {
//	ContractId   []byte             `json:"contract_id"`   // contract id
//	Args         [][]byte           `json:"args"`          // contract arguments list
//	Excutiontime time.Duration      `json:"excution_time"` // contract execution time, millisecond
//	ReadSet      []ContractReadSet  `json:"read_set"`      // the set data of read, and value could be any type
//	WriteSet     []PayloadMapStruct `json:"write_set"`     // the set data of write, and value could be any type
//	Payload      []byte             `json:"payload"`
//}

// Token exchange message and verify message
// App: config	// update global config
type ConfigPayload struct {
	ConfigSet []PayloadMapStruct `json:"config_set"` // the array of global config
}

// Token exchange message and verify message
// App: text
type TextPayload struct {
	Text []byte `json:"text"` // Textdata
}

/************************** End of Payload Details ******************************************/

type Author struct {
	Address        common.Address `json:"address"`
	Pubkey         []byte/*common.Hash*/ `json:"pubkey"`
	TxAuthentifier Authentifier `json:"authentifiers"`
}

type Authentifier struct {
	Address string `json:"address"`
	R       []byte `json:"r"`
	S       []byte `json:"s"`
	V       []byte `json:"v"`
}



//func CopyTransactions(txs Transactions) Transactions {
//	cpy := txs
//	return cpy
//}

type UnitNonce [8]byte

/************************** Unit Members  *****************************/
func (u *Unit) Header() *Header { return u.UnitHeader }

// transactions
func (u *Unit) Transactions() []*Transaction {
	return u.Txs
}

// return transaction
func (u *Unit) Transaction(hash common.Hash) *Transaction {
	for _, transaction := range u.Txs {
		if transaction.TxHash == hash {
			return transaction
		}
	}
	return nil
}

// return  unit'UnitHash




// return Creationdate
// comment by Albert·Gou
//func (u *Unit) CreationDate() time.Time {
//	return u.UnitHeader.Creationdate
//}

//func (u *Unit) NumberU64() uint64 { return u.Head.Number.Uint64() }
func (u *Unit) Number() ChainIndex {
	return u.UnitHeader.Number
}
func (u *Unit) NumberU64() uint64 {
	return u.UnitHeader.Number.Index
}

// return unit's parents UnitHash
func (u *Unit) ParentHash() []common.Hash {
	return u.UnitHeader.ParentsHash
}

type ErrUnit float64

func (e ErrUnit) Error() string {
	switch e {
	case -1:
		return "Unit size error"
	case -2:
		return "Unit signature error"
	case -3:
		return "Unit header save error"
	case -4:
		return "Unit tx size error"
	case -5:
		return "Save create token transaction error"
	case -6:
		return "Save config transaction error"
	default:
		return ""
	}
	return ""
}

/************************** Unit Members  *****************************/

// NewBlockWithHeader creates a block with the given header data. The
// header data is copied, changes to header and to the field values
// will not affect the block.
func NewUnitWithHeader(header *Header) *Unit {
	return &Unit{UnitHeader: header}
}

// WithBody returns a new block with the given transaction and uncle contents.
func (b *Unit) WithBody(transactions *Transactions) *Unit {
	root := core.DeriveSha(transactions)
	if strings.Compare(root.String(), b.UnitHeader.TxRoot.String()) != 0 {
		return nil
	}
	// set unit body
	b.Txs = *transactions
	return b
}

func (u *Unit) ContainsParent(pHash common.Hash) bool {
	ps := pHash.String()
	for _, hash := range u.UnitHeader.ParentsHash {
		if strings.Compare(hash.String(), ps) == 0 {
			return true
		}
	}
	return false
}

//func RSVtoAddress(tx *Transaction) common.Address {
//	//sig := make([]byte, 65)
//	//copy(sig[32-len(tx.From.R):32], tx.From.R)
//	//copy(sig[64-len(tx.From.S):64], tx.From.S)
//	//copy(sig[64:], tx.From.V)
//	//pub, _ := crypto.SigToPub(tx.TxHash[:], sig)
//	//address := crypto.PubkeyToAddress(*pub)
//	//return address
//	return common.Address{}
//}
//
//func MsgstoAddress(msgs []Message) common.Address {
//	forms := make([]common.Address, 0)
//	//payment load to address.
//
//	for _, msg := range msgs {
//		payment, ok := msg.Payload.(PaymentPayload)
//		if !ok {
//			break
//		}
//		for _, pay := range payment.Inputs {
//			// 通过签名信息还原出address
//			from := new(common.Address)
//			from.SetBytes(pay.Extra[:])
//			forms = append(forms, *from)
//		}
//	}
//	if len(forms) > 0 {
//		return forms[0]
//	}
//	return common.Address{}
//}
//func RSVtoPublicKey(hash, r, s, v []byte) (*ecdsa.PublicKey, error) {
//	sig := make([]byte, 65)
//	copy(sig[32-len(r):32], r)
//	copy(sig[64-len(s):64], s)
//	copy(sig[64:], v)
//	return crypto.SigToPub(hash, sig)
//}

type TxValidationCode int32

const (
	TxValidationCode_VALID                        TxValidationCode = 0
	TxValidationCode_INVALID_CONTRACT_TEMPLATE    TxValidationCode = 1
	TxValidationCode_INVALID_FEE                  TxValidationCode = 2
	TxValidationCode_BAD_COMMON_HEADER            TxValidationCode = 3
	TxValidationCode_BAD_CREATOR_SIGNATURE        TxValidationCode = 4
	TxValidationCode_INVALID_ENDORSER_TRANSACTION TxValidationCode = 5
	TxValidationCode_INVALID_CONFIG_TRANSACTION   TxValidationCode = 6
	TxValidationCode_UNSUPPORTED_TX_PAYLOAD       TxValidationCode = 7
	TxValidationCode_BAD_PROPOSAL_TXID            TxValidationCode = 8
	TxValidationCode_DUPLICATE_TXID               TxValidationCode = 9
	TxValidationCode_ENDORSEMENT_POLICY_FAILURE   TxValidationCode = 10
	TxValidationCode_MVCC_READ_CONFLICT           TxValidationCode = 11
	TxValidationCode_PHANTOM_READ_CONFLICT        TxValidationCode = 12
	TxValidationCode_UNKNOWN_TX_TYPE              TxValidationCode = 13
	TxValidationCode_TARGET_CHAIN_NOT_FOUND       TxValidationCode = 14
	TxValidationCode_MARSHAL_TX_ERROR             TxValidationCode = 15
	TxValidationCode_NIL_TXACTION                 TxValidationCode = 16
	TxValidationCode_EXPIRED_CHAINCODE            TxValidationCode = 17
	TxValidationCode_CHAINCODE_VERSION_CONFLICT   TxValidationCode = 18
	TxValidationCode_BAD_HEADER_EXTENSION         TxValidationCode = 19
	TxValidationCode_BAD_CHANNEL_HEADER           TxValidationCode = 20
	TxValidationCode_BAD_RESPONSE_PAYLOAD         TxValidationCode = 21
	TxValidationCode_BAD_RWSET                    TxValidationCode = 22
	TxValidationCode_ILLEGAL_WRITESET             TxValidationCode = 23
	TxValidationCode_INVALID_WRITESET             TxValidationCode = 24
	TxValidationCode_NOT_VALIDATED                TxValidationCode = 254
	TxValidationCode_NOT_COMPARE_SIZE             TxValidationCode = 255
	TxValidationCode_INVALID_OTHER_REASON         TxValidationCode = 256
)

var TxValidationCode_name = map[int32]string{
	0:   "VALID",
	1:   "INVALID_CONTRACT_TEMPLATE",
	2:   "INVALID_FEE",
	3:   "BAD_COMMON_HEADER",
	4:   "BAD_CREATOR_SIGNATURE",
	5:   "INVALID_ENDORSER_TRANSACTION",
	6:   "INVALID_CONFIG_TRANSACTION",
	7:   "UNSUPPORTED_TX_PAYLOAD",
	8:   "BAD_PROPOSAL_TXID",
	9:   "DUPLICATE_TXID",
	10:  "ENDORSEMENT_POLICY_FAILURE",
	11:  "MVCC_READ_CONFLICT",
	12:  "PHANTOM_READ_CONFLICT",
	13:  "UNKNOWN_TX_TYPE",
	14:  "TARGET_CHAIN_NOT_FOUND",
	15:  "MARSHAL_TX_ERROR",
	16:  "NIL_TXACTION",
	17:  "EXPIRED_CHAINCODE",
	18:  "CHAINCODE_VERSION_CONFLICT",
	19:  "BAD_HEADER_EXTENSION",
	20:  "BAD_CHANNEL_HEADER",
	21:  "BAD_RESPONSE_PAYLOAD",
	22:  "BAD_RWSET",
	23:  "ILLEGAL_WRITESET",
	24:  "INVALID_WRITESET",
	254: "NOT_VALIDATED",
	255: "NOT_COMPARE_SIZE",
	256: "INVALID_OTHER_REASON",
}

/**
根据大端规则填充字节
To full fill bytes according bigendian
*/
//func FillBytes(data []byte, lenth uint8) []byte {
//	newBytes := make([]byte, lenth)
//	if len(data) < int(lenth) {
//		len := int(lenth) - len(data)
//		for i, b := range data {
//			newBytes[len+i] = b
//		}
//	} else {
//		newBytes = data[:lenth]
//	}
//	return newBytes
//}