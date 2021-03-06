// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// This file contains some shares testing functionality, common to  multiple
// different files and modules being tested.

package ptn

import (
	"crypto/ecdsa"
	"crypto/rand"
	//"math/big"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/palletone/go-palletone/common"
	//"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/consensus"

	//"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"

	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/consensus/mediatorplugin"
	common2 "github.com/palletone/go-palletone/dag/common"
	"time"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/dag"
)

//var (
//	testBankKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
//	testBank       = crypto.PubkeyToAddress(testBankKey.PublicKey)
//)

// newTestProtocolManager creates a new protocol manager for testing purposes,
// with the given number of blocks already known, and potential notification
// channels for different events.
func newTestProtocolManager(mode downloader.SyncMode, blocks int, newtx chan<- []*modules.Transaction) (*ProtocolManager, ptndb.Database, error) {
	memdb, _ := ptndb.NewMemDatabase()
	dag, _ := MakeDags(memdb, blocks)
	//uu := dag.CurrentUnit()
	//log.Printf("--------newTestProtocolManager--CurrentUnit--unit.UnitHeader-----%#v\n", uu.UnitHeader)
	//log.Printf("--------newTestProtocolManager--CurrentUnit--unit.UnitHash-------%#v\n", uu.UnitHash)
	//log.Printf("--------newTestProtocolManager--CurrentUnit--unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
	//log.Printf("--------newTestProtocolManager--CurrentUnit--unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	//index := modules.ChainIndex{
	//	modules.PTNCOIN,
	//	true,
	//	0,
	//}
	//uu = dag.GetUnitByNumber(index)
	//log.Printf("--------newTestProtocolManager--index=0--unit.UnitHeader-----%#v\n", uu.UnitHeader)
	//log.Printf("--------newTestProtocolManager--index=0--unit.UnitHash-------%#v\n", uu.UnitHash)
	//log.Printf("--------newTestProtocolManager--index=0--unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
	//log.Printf("--------newTestProtocolManager--index=0--unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	engine := new(consensus.DPOSEngine)
	typemux := new(event.TypeMux)
	producer := new(mediatorplugin.MediatorPlugin)
	pm, err := NewProtocolManager(mode, DefaultConfig.NetworkId, &testTxPool{added: newtx},
		engine, dag, typemux, memdb, producer)
	if err != nil {
		return nil, nil, err
	}
	pm.Start(1000)
	return pm, memdb, nil
}

// newTestProtocolManagerMust creates a new protocol manager for testing purposes,
// with the given number of blocks already known, and potential notification
// channels for different events. In case of an error, the constructor force-
// fails the test.
func newTestProtocolManagerMust(t *testing.T, mode downloader.SyncMode, blocks int, newtx chan<- []*modules.Transaction) (*ProtocolManager, ptndb.Database) {
	pm, db, err := newTestProtocolManager(mode, blocks /*generator,*/, newtx)
	if err != nil {
		t.Fatalf("Failed to create protocol manager: %v", err)
	}
	return pm, db
}

// testTxPool is a fake, helper transaction pool for testing purposes
type testTxPool struct {
	txFeed event.Feed
	pool   []*modules.Transaction        // Collection of all transactions
	added  chan<- []*modules.Transaction // Notification channel for new transactions

	lock sync.RWMutex // Protects the transaction pool
}

// AddRemotes appends a batch of transactions to the pool, and notifies any
// listeners if the addition channel is non nil
func (p *testTxPool) AddRemotes(txs []*modules.Transaction) []error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.pool = append(p.pool, txs...)
	if p.added != nil {
		p.added <- txs
	}
	return make([]error, len(txs))
}

// Pending returns all the transactions known to the pool
func (p *testTxPool) Pending() (map[common.Hash]*modules.TxPoolTransaction, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	batches := make(map[common.Hash]*modules.TxPoolTransaction)
	//for _, tx := range p.pool {
	// from, _ := types.Sender(types.HomesteadSigner{}, tx)
	// batches[from] = append(batches[from], tx)
	//}
	//for _, batch := range batches {
	// sort.Sort(types.TxByNonce(batch))
	//}
	return batches, nil
}

func (p *testTxPool) SubscribeTxPreEvent(ch chan<- modules.TxPreEvent) event.Subscription {
	return p.txFeed.Subscribe(ch)
}

// newTestTransaction create a new dummy transaction.
func newTestTransaction(from *ecdsa.PrivateKey, nonce uint64, datasize int) *modules.Transaction {
	msg := &modules.Message{
		App: modules.APP_PAYMENT,
		//PayloadHash: common.HexToHash("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"),
		Payload: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	//tx := modules.NewTransaction(nonce, big.NewInt(0), []byte("abc"))
	tx := modules.NewTransaction(
		[]*modules.Message{msg, msg, msg},
		12345,
	)

	return tx
}

// testPeer is a simulated peer to allow testing direct network calls.
type testPeer struct {
	net p2p.MsgReadWriter // Network layer reader/writer to simulate remote messaging
	app *p2p.MsgPipeRW    // Application layer reader/writer to simulate the local side
	*peer
}

// newTestPeer creates a new peer registered at the given protocol manager.
func newTestPeer(name string, version int, pm *ProtocolManager, shake bool) (*testPeer, <-chan error) {
	// Create a message pipe to communicate through
	app, net := p2p.MsgPipe()

	// Generate a random id and create the peer
	var id discover.NodeID
	rand.Read(id[:])

	peer := pm.newPeer(version, p2p.NewPeer(id, name, nil), net)

	// Start the peer on a new thread
	errc := make(chan error, 1)
	go func() {
		select {
		case pm.newPeerCh <- peer:
			errc <- pm.handle(peer)
		case <-pm.quitSync:
			errc <- p2p.DiscQuitting
		}
	}()
	tp := &testPeer{app: app, net: net, peer: peer}
	// Execute any implicitly requested handshakes and return
	if shake {
		var (
			//number = modules.ChainIndex{
			//	modules.PTNCOIN,
			//	true,
			//	0,
			//}
			//genesis = pm.dag.GetUnitByNumber(number)
			head = pm.dag.CurrentHeader()
			td   = head.Number.Index
		)
		//fmt.Println("	if shake {===》》》",td)
		genesis, err := common2.GetGenesisUnit(pm.dag.Db, 0)
		//fmt.Println("genesis unti if shake {===》》》",genesis.UnitHash)
		if err != nil {
			fmt.Println("GetGenesisUnit===error:=", err)
		}
		tp.handshake(nil, td, head.Hash(), genesis.Hash())
	}
	return tp, errc
}

// handshake simulates a trivial handshake that expects the same state from the
// remote side as we are simulating locally.
func (p *testPeer) handshake(t *testing.T, td uint64, head common.Hash, genesis common.Hash) {
	msg := &statusData{
		ProtocolVersion: uint32(p.version),
		NetworkId:       DefaultConfig.NetworkId,
		TD:              td,
		CurrentBlock:    head,
		GenesisBlock:    genesis,
	}
	if err := p2p.ExpectMsg(p.app, StatusMsg, msg); err != nil {
		log.Fatalf("status recv: %v", err)
	}
	if err := p2p.Send(p.app, StatusMsg, msg); err != nil {
		log.Fatalf("status send: %v", err)
	}
}

// close terminates the local side of the peer, notifying the remote protocol
// manager of termination.
func (p *testPeer) close() {
	p.app.Close()
}

func MakeDags(Memdb ptndb.Database, unitAccount int) (*dag.Dag, error) {
	dag, _ := dag.NewDagForTest(Memdb)
	header := NewHeader([]common.Hash{}, []modules.IDType16{modules.PTNCOIN}, []byte{})
	header.Number.AssetID = modules.PTNCOIN
	header.Number.IsMain = true
	header.Number.Index = 0
	header.Authors = &modules.Authentifier{"", []byte{}, []byte{}, []byte{}}
	header.Witness = []*modules.Authentifier{&modules.Authentifier{"", []byte{}, []byte{}, []byte{}}}
	tx, _ := NewCoinbaseTransaction()
	txs := modules.Transactions{tx}
	genesisUnit := NewUnit(header, txs)
	err := SaveGenesis(dag.Db, genesisUnit)
	if err != nil {
		log.Println("SaveGenesis, err", err)
		return nil, err
	}
	newDag(dag.Db, genesisUnit, unitAccount)
	return dag, nil
}
func newDag(memdb ptndb.Database, gunit *modules.Unit, number int) (modules.Units, error) {
	units := make(modules.Units, number)
	par := gunit
	for i := 0; i < number; i++ {
		header := NewHeader([]common.Hash{par.UnitHash}, []modules.IDType16{modules.PTNCOIN}, []byte{})
		header.Number.AssetID = par.UnitHeader.Number.AssetID
		header.Number.IsMain = par.UnitHeader.Number.IsMain
		header.Number.Index = par.UnitHeader.Number.Index + 1
		header.Authors = &modules.Authentifier{"", []byte{}, []byte{}, []byte{}}
		header.Witness = []*modules.Authentifier{&modules.Authentifier{"", []byte{}, []byte{}, []byte{}}}
		tx, _ := NewCoinbaseTransaction()
		txs := modules.Transactions{tx}
		unit := NewUnit(header, txs)
		err := SaveUnit(memdb, unit, true)
		if err != nil {
			log.Println("save genesis error", err)
			return nil, err
		}
		units[i] = unit
		par = unit
	}
	return units, nil
}
func SaveGenesis(db ptndb.Database, unit *modules.Unit) error {
	if unit.NumberU64() != 0 {
		return fmt.Errorf("can't commit genesis unit with number > 0")
	}
	err := SaveUnit(db, unit, true)
	if err != nil {
		log.Println("SaveGenesis==", err)
		return err
	}
	return nil
}

func SaveUnit(db ptndb.Database, unit *modules.Unit, isGenesis bool) error {
	if unit.UnitSize == 0 || unit.Size() == 0 {
		log.Println("Unit is null")
		return fmt.Errorf("Unit is null")
	}
	if unit.UnitSize != unit.Size() {
		log.Println("Validate size", "error", "Size is invalid")
		return modules.ErrUnit(-1)
	}
	_, isSuccess, err := common2.ValidateTransactions(db, &unit.Txs, isGenesis)
	if isSuccess != true {
		fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
		return fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
	}
	// step4. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	if err := storage.SaveHeader(db, unit.UnitHash, unit.UnitHeader); err != nil {
		log.Println("SaveHeader:", "error", err.Error())
		return modules.ErrUnit(-3)
	}
	// step5. save unit hash and chain index relation
	// key is like "[UNIT_HASH_NUMBER][unit_hash]"
	if err := storage.SaveNumberByHash(db, unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Println("SaveHashNumber:", "error", err.Error())
		return fmt.Errorf("Save unit hash and number error")
	}
	if err := storage.SaveHashByNumber(db, unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Println("SaveNumberByHash:", "error", err.Error())
		return fmt.Errorf("Save unit hash and number error")
	}
	if err := storage.SaveTxLookupEntry(db, unit); err != nil {
		return err
	}
	if err := storage.SaveTxLookupEntry(db, unit); err != nil {
		return err
	}
	if err := saveHashByIndex(db, unit.UnitHash, unit.UnitHeader.Number.Index); err != nil {
		return err
	}
	// update state
	storage.PutCanonicalHash(db, unit.UnitHash, unit.NumberU64())
	storage.PutHeadHeaderHash(db, unit.UnitHash)
	storage.PutHeadUnitHash(db, unit.UnitHash)
	storage.PutHeadFastUnitHash(db, unit.UnitHash)
	// todo send message to transaction pool to delete unit's transactions
	return nil
}
func NewUnit(header *modules.Header, txs modules.Transactions) *modules.Unit {
	u := &modules.Unit{
		UnitHeader: header,
		Txs:        txs,
	}
	u.ReceivedAt = time.Now()
	u.UnitSize = u.Size()
	u.UnitHash = u.Hash()
	return u
}
func NewHeader(parents []common.Hash, asset []modules.IDType16, extra []byte) *modules.Header {
	hashs := make([]common.Hash, 0)
	hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	var b []byte
	return &modules.Header{ParentsHash: hashs, AssetIDs: asset, Extra: append(b, extra...), Creationdate: time.Now().Unix()}
}
func NewCoinbaseTransaction() (*modules.Transaction, error) {
	input := &modules.Input{}
	output := &modules.Output{}
	payload := modules.PaymentPayload{
		Input:  []*modules.Input{input},
		Output: []*modules.Output{output},
	}
	msg := modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: payload,
	}
	var coinbase modules.Transaction
	coinbase.TxMessages = append(coinbase.TxMessages, &msg)
	coinbase.TxHash = coinbase.Hash()
	return &coinbase, nil
}

func saveHashByIndex(db ptndb.Database, hash common.Hash, index uint64) error {
	key := fmt.Sprintf("%s%v_", storage.HEADER_PREFIX, index)
	err := db.Put([]byte(key), hash.Bytes())
	return err
}

