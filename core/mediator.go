/*
	This file is part of go-palletone.
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
*/
/*
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package core

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/pairing/bn256"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
)

// mediator 结构体 和具体的账户模型有关
type Mediator struct {
	Address     common.Address
	InitPartPub kyber.Point
	Node        *discover.Node
}

func StrToMedNode(mn string) *discover.Node {
	node, err := discover.ParseNode(mn)
	if err != nil {
		log.Error(fmt.Sprintf("Invalid mediator node \"%v\" : %v", mn, err))
	}

	return node
}

func StrToMedAdd(addStr string) common.Address {
	address := strings.TrimSpace(addStr)
	address = strings.Trim(address, "\"")

	addr, err := common.StringToAddress(address)
	// addrType, err := addr.Validate()
	if err != nil || addr.GetType() != common.PublicKeyHash {
		log.Error(fmt.Sprintf("Invalid mediator account address \"%v\" : %v", address, err))
	}

	return addr
}

func StrToScalar(secStr string) kyber.Scalar {
	secB, err := base64.RawURLEncoding.DecodeString(secStr)
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	sec := bn256.NewSuiteG2().Scalar()
	err = sec.UnmarshalBinary(secB)
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	return sec
}

func StrToPoint(pubStr string) kyber.Point {
	pubB, err := base64.RawURLEncoding.DecodeString(pubStr)
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	pub := bn256.NewSuiteG2().Point()
	err = pub.UnmarshalBinary(pubB)
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	return pub
}

func InfoToMediator(medInfo *MediatorInfo) Mediator {
	// 1. 解析 mediator 账户地址
	add := StrToMedAdd(medInfo.Address)

	// 2. 解析 mediator 的 DKS 初始公钥
	pub := StrToPoint(medInfo.InitPartPub)

	// 3. 解析mediator 的 node 节点信息
	node := StrToMedNode(medInfo.Node)

	md := Mediator{
		Address:     add,
		InitPartPub: pub,
		Node:        node,
	}

	return md
}
