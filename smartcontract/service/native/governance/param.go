/*
 * Copyright (C) 2018 The ZeepinChain Authors
 * This file is part of The ZeepinChain library.
 *
 * The ZeepinChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ZeepinChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ZeepinChain.  If not, see <http://www.gnu.org/licenses/>.

 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package governance

import (
	"io"

	"math"

	"github.com/imZhuFei/zeepin/common"
	"github.com/imZhuFei/zeepin/common/serialization"
	"github.com/imZhuFei/zeepin/errors"
	"github.com/imZhuFei/zeepin/smartcontract/service/native/utils"
)

type RegisterCandidateParam struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint64
	Caller     []byte
	KeyNo      uint32
}

func (this *RegisterCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, request peerPubkey error!")
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, address address error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.InitPos)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize initPos error!")
	}
	if err := serialization.WriteVarBytes(w, this.Caller); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, serialize caller error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.KeyNo)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize keyNo error!")
	}
	return nil
}

func (this *RegisterCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadAddress, deserialize address error!")
	}
	initPos, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize initPos error!")
	}
	if initPos > math.MaxInt64 {
		return errors.NewErr("initPos larger than max of uint64!")
	}
	caller, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarBytes, deserialize caller error!")
	}
	keyNo, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize keyNo error!")
	}
	if keyNo > math.MaxUint32 {
		return errors.NewErr("initPos larger than max of uint32!")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.InitPos = initPos
	this.Caller = caller
	this.KeyNo = uint32(keyNo)
	return nil
}

type UnRegisterCandidateParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *UnRegisterCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, request peerPubkey error!")
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, address address error!")
	}
	return nil
}

func (this *UnRegisterCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadAddress, deserialize address error!")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type GetVoteInfoParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *GetVoteInfoParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize peerPubkey error!")
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, address address error!")
	}
	return nil
}

func (this *GetVoteInfoParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadAddress, deserialize address error!")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type QuitNodeParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *QuitNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize peerPubkey error!")
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, address address error!")
	}
	return nil
}

func (this *QuitNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadAddress, deserialize address error!")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type ApproveCandidateParam struct {
	PeerPubkey string
}

func (this *ApproveCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	return nil
}

func (this *ApproveCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type RejectCandidateParam struct {
	PeerPubkey string
}

func (this *RejectCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	return nil
}

func (this *RejectCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type CheckVoteInfoParam struct {
	PeerPubkey string
}

func (this *CheckVoteInfoParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	return nil
}

func (this *CheckVoteInfoParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type BlackNodeParam struct {
	PeerPubkeyList []string
}

func (this *BlackNodeParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(len(this.PeerPubkeyList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize peerPubkeyList length error!")
	}
	for _, v := range this.PeerPubkeyList {
		if err := serialization.WriteString(w, v); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
		}
	}
	return nil
}

func (this *BlackNodeParam) Deserialize(r io.Reader) error {
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize peerPubkeyList length error!")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	this.PeerPubkeyList = peerPubkeyList
	return nil
}

type WhiteNodeParam struct {
	PeerPubkey string
}

func (this *WhiteNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	return nil
}

func (this *WhiteNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type VoteForPeerParam struct {
	Address        common.Address
	PeerPubkeyList []string
	PosList        []uint64
}

func (this *VoteForPeerParam) Serialize(w io.Writer) error {
	if len(this.PeerPubkeyList) > 1024 {
		return errors.NewErr("length of input list > 1024!")
	}
	if len(this.PeerPubkeyList) != len(this.PosList) {
		return errors.NewErr("length of PeerPubkeyList != length of PosList!")
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, address address error!")
	}
	if err := utils.WriteVarUint(w, uint64(len(this.PeerPubkeyList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize peerPubkeyList length error!")
	}
	for _, v := range this.PeerPubkeyList {
		if err := serialization.WriteString(w, v); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
		}
	}
	if err := utils.WriteVarUint(w, uint64(len(this.PosList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize posList length error!")
	}
	for _, v := range this.PosList {
		if err := utils.WriteVarUint(w, uint64(v)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize pos error!")
		}
	}
	return nil
}

func (this *VoteForPeerParam) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadAddress, deserialize address error!")
	}
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize peerPubkeyList length error!")
	}
	if n > 1024 {
		return errors.NewErr("length of input list > 1024!")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	m, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize posList length error!")
	}
	posList := make([]uint64, 0)
	for i := 0; uint64(i) < m; i++ {
		k, err := utils.ReadVarUint(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize pos error!")
		}
		if k > math.MaxInt64 {
			return errors.NewErr("pos larger than max of uint64!")
		}
		posList = append(posList, k)
	}
	if m != n {
		return errors.NewErr("length of PeerPubkeyList != length of PosList!")
	}
	this.Address = address
	this.PeerPubkeyList = peerPubkeyList
	this.PosList = posList
	return nil
}

type WithdrawParam struct {
	Address        common.Address
	PeerPubkeyList []string
	WithdrawList   []uint64
}

func (this *WithdrawParam) Serialize(w io.Writer) error {
	if len(this.PeerPubkeyList) > 1024 {
		return errors.NewErr("length of input list > 1024!")
	}
	if len(this.PeerPubkeyList) != len(this.WithdrawList) {
		return errors.NewErr("length of PeerPubkeyList != length of WithdrawList, contract params error!")
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, address address error!")
	}
	if err := utils.WriteVarUint(w, uint64(len(this.PeerPubkeyList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize peerPubkeyList length error!")
	}
	for _, v := range this.PeerPubkeyList {
		if err := serialization.WriteString(w, v); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
		}
	}
	if err := utils.WriteVarUint(w, uint64(len(this.WithdrawList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize withdrawList length error!")
	}
	for _, v := range this.WithdrawList {
		if err := utils.WriteVarUint(w, uint64(v)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize withdraw error!")
		}
	}
	return nil
}

func (this *WithdrawParam) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadAddress, deserialize address error!")
	}
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize peerPubkeyList length error!")
	}
	if n > 1024 {
		return errors.NewErr("length of input list > 1024!")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	m, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize withdrawList length error!")
	}
	withdrawList := make([]uint64, 0)
	for i := 0; uint64(i) < m; i++ {
		k, err := utils.ReadVarUint(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize withdraw error!")
		}
		if k > math.MaxInt64 {
			return errors.NewErr("pos larger than max of uint64!")
		}
		withdrawList = append(withdrawList, k)
	}
	if m != n {
		return errors.NewErr("length of PeerPubkeyList != length of WithdrawList, contract params error!")
	}
	this.Address = address
	this.PeerPubkeyList = peerPubkeyList
	this.WithdrawList = withdrawList
	return nil
}

type Configuration struct {
	N                    uint32
	C                    uint32
	K                    uint32
	L                    uint32
	BlockMsgDelay        uint32
	HashMsgDelay         uint32
	PeerHandshakeTimeout uint32
	MaxBlockChangeView   uint32
}

func (this *Configuration) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.N)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize n error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.C)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize c error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.K)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize k error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.L)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize l error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.BlockMsgDelay)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize block_msg_delay error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.HashMsgDelay)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize hash_msg_delay error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.PeerHandshakeTimeout)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize peer_handshake_timeout error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.MaxBlockChangeView)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize max_block_change_view error!")
	}
	return nil
}

func (this *Configuration) Deserialize(r io.Reader) error {
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize n error!")
	}
	c, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize c error!")
	}
	k, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize k error!")
	}
	l, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize l error!")
	}
	blockMsgDelay, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize blockMsgDelay error!")
	}
	hashMsgDelay, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize hashMsgDelay error!")
	}
	peerHandshakeTimeout, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize peerHandshakeTimeout error!")
	}
	maxBlockChangeView, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize maxBlockChangeView error!")
	}
	if n > math.MaxUint32 {
		return errors.NewErr("n larger than max of uint32!")
	}
	if c > math.MaxUint32 {
		return errors.NewErr("c larger than max of uint32!")
	}
	if k > math.MaxUint32 {
		return errors.NewErr("k larger than max of uint32!")
	}
	if l > math.MaxUint32 {
		return errors.NewErr("l larger than max of uint32!")
	}
	if blockMsgDelay > math.MaxUint32 {
		return errors.NewErr("blockMsgDelay larger than max of uint32!")
	}
	if hashMsgDelay > math.MaxUint32 {
		return errors.NewErr("hashMsgDelay larger than max of uint32!")
	}
	if peerHandshakeTimeout > math.MaxUint32 {
		return errors.NewErr("peerHandshakeTimeout larger than max of uint32!")
	}
	if maxBlockChangeView > math.MaxUint32 {
		return errors.NewErr("maxBlockChangeView larger than max of uint32!")
	}
	this.N = uint32(n)
	this.C = uint32(c)
	this.K = uint32(k)
	this.L = uint32(l)
	this.BlockMsgDelay = uint32(blockMsgDelay)
	this.HashMsgDelay = uint32(hashMsgDelay)
	this.PeerHandshakeTimeout = uint32(peerHandshakeTimeout)
	this.MaxBlockChangeView = uint32(maxBlockChangeView)
	return nil
}

type GlobalParam struct {
	CandidateFee uint64 //unit: 10^-4 gala
	MinInitStake uint64
	CandidateNum uint32
	PosLimit     uint32
	A            uint32
	B            uint32
	Yita         uint32
	Penalty      uint32
}

func (this *GlobalParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.CandidateFee); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize candidateFee error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.MinInitStake)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize minInitStake error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.CandidateNum)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize candidateNum error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.PosLimit)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize posLimit error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.A)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize a error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.B)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize b error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.Yita)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize yita error!")
	}
	if err := utils.WriteVarUint(w, uint64(this.Penalty)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize penalty error!")
	}
	return nil
}

func (this *GlobalParam) Deserialize(r io.Reader) error {
	candidateFee, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize candidateFee error!")
	}
	minInitStake, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize minInitStake error!")
	}
	candidateNum, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize candidateNum error!")
	}
	posLimit, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize posLimit error!")
	}
	a, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize a error!")
	}
	b, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize b error!")
	}
	yita, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize yita error!")
	}
	penalty, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize penalty error!")
	}
	if minInitStake > math.MaxUint64 {
		return errors.NewErr("minInitStake larger than max of uint32!")
	}
	if candidateNum > math.MaxUint32 {
		return errors.NewErr("candidateNum larger than max of uint32!")
	}
	if posLimit > math.MaxUint32 {
		return errors.NewErr("posLimit larger than max of uint32!")
	}
	if a > math.MaxUint32 {
		return errors.NewErr("a larger than max of uint32!")
	}
	if b > math.MaxUint32 {
		return errors.NewErr("b larger than max of uint32!")
	}
	if yita > math.MaxUint32 {
		return errors.NewErr("yita larger than max of uint32!")
	}
	if penalty > math.MaxUint32 {
		return errors.NewErr("penalty larger than max of uint32!")
	}
	this.CandidateFee = candidateFee
	this.MinInitStake = uint64(minInitStake)
	this.CandidateNum = uint32(candidateNum)
	this.PosLimit = uint32(posLimit)
	this.A = uint32(a)
	this.B = uint32(b)
	this.Yita = uint32(yita)
	this.Penalty = uint32(penalty)
	return nil
}

type SplitCurve struct {
	Yi []uint32
}

func (this *SplitCurve) Serialize(w io.Writer) error {
	if len(this.Yi) != 101 {
		return errors.NewErr("length of split curve != 101!")
	}
	if err := utils.WriteVarUint(w, uint64(len(this.Yi))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize Yi length error!")
	}
	for _, v := range this.Yi {
		if err := utils.WriteVarUint(w, uint64(v)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "utils.WriteVarUint, serialize splitCurve error!")
		}
	}
	return nil
}

func (this *SplitCurve) Deserialize(r io.Reader) error {
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize Yi length error!")
	}
	yi := make([]uint32, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := utils.ReadVarUint(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadVarUint, deserialize splitCurve error!")
		}
		if k > math.MaxUint32 {
			return errors.NewErr("yi larger than max of uint32!")
		}
		yi = append(yi, uint32(k))
	}
	this.Yi = yi
	return nil
}

type TransferPenaltyParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *TransferPenaltyParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize peerPubkey error!")
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, address address error!")
	}
	return nil
}

func (this *TransferPenaltyParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadAddress, deserialize address error!")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type WithdrawGalaParam struct {
	Address common.Address
}

func (this *WithdrawGalaParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, address address error!")
	}
	return nil
}

func (this *WithdrawGalaParam) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "utils.ReadAddress, deserialize address error!")
	}
	this.Address = address
	return nil
}
