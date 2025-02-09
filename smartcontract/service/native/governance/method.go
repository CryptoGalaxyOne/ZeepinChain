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
	"bytes"
	"encoding/hex"
	"sort"

	"github.com/imZhuFei/zeepin/common"
	"github.com/imZhuFei/zeepin/common/constants"
	"github.com/imZhuFei/zeepin/common/log"
	cstates "github.com/imZhuFei/zeepin/core/states"
	scommon "github.com/imZhuFei/zeepin/core/store/common"
	"github.com/imZhuFei/zeepin/errors"
	"github.com/imZhuFei/zeepin/smartcontract/service/native"
	"github.com/imZhuFei/zeepin/smartcontract/service/native/utils"
)

func registerCandidate(native *native.NativeService, flag string) error {
	params := new(RegisterCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check auth of GID
	//err := appCallVerifyToken(native, contract, params.Caller, REGISTER_CANDIDATE, uint64(params.KeyNo))
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "appCallVerifyToken, verifyToken failed!")
	//}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//check peerPubkey
	if err := validatePeerPubKeyFormat(params.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "invalid peer pubkey")
	}

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}

	//get black list
	blackList, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get BlackList error!")
	}
	if blackList != nil {
		return errors.NewErr("registerCandidate, this Peer is in BlackList!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//check if exist in PeerPool
	_, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if ok {
		return errors.NewErr("registerCandidate, peerPubkey is already in peerPoolMap!")
	}
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}
	//check initPos
	if uint64(params.InitPos) < globalParam.MinInitStake {
		return errors.NewDetailErr(err, errors.ErrNoCode, "initPos must >=MinInitStake")
	}
	peerPoolItem := &PeerPoolItem{
		PeerPubkey: params.PeerPubkey,
		Address:    params.Address,
		InitPos:    uint64(params.InitPos),
		TotalPos:   0,
		Status:     RegisterCandidateStatus,
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}
	log.Infof("registNode: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, peerPoolItem.PeerPubkey)
	switch flag {
	case "transfer":
		//zpt transfer
		err = appCallTransferZpt(native, params.Address, utils.GovernanceContractAddress, uint64(params.InitPos))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferZpt, zpt transfer error!")
		}

		//gala transfer
		err = appCallTransferGala(native, params.Address, utils.GovernanceContractAddress, globalParam.CandidateFee)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferGala, gala transfer error!")
		}
	case "transferFrom":
		//zpt transfer from
		err = appCallTransferFromZpt(native, utils.GovernanceContractAddress, params.Address, utils.GovernanceContractAddress, uint64(params.InitPos))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromZpt, zpt transfer error!")
		}

		//gala transfer from
		err = appCallTransferFromGala(native, utils.GovernanceContractAddress, params.Address, utils.GovernanceContractAddress, globalParam.CandidateFee)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromGala, gala transfer error!")
		}
	}

	//update total stake
	err = depositTotalStake(native, contract, params.Address, uint64(params.InitPos))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "depositTotalStake, depositTotalStake error!")
	}
	return nil
}

func voteForPeer(native *native.NativeService, flag string) error {
	params := &VoteForPeerParam{
		PeerPubkeyList: make([]string, 0),
		PosList:        make([]uint64, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	var total uint64
	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := params.PosList[i]

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus && peerPoolItem.Status != RegisterCandidateStatus {
			return errors.NewErr("voteForPeer, peerPubkey is not candidate and can not be voted!")
		}

		voteInfo, err := getVoteInfo(native, contract, peerPubkey, params.Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "getVoteInfo, get voteInfo error!")
		}
		if pos < 0 {
			return errors.NewDetailErr(err, errors.ErrNoCode, "vote pos must greater than zero!")
		}
		voteInfo.NewPos = voteInfo.NewPos + uint64(pos)
		total = total + uint64(pos)
		peerPoolItem.TotalPos = peerPoolItem.TotalPos + uint64(pos)
		log.Infof("voteForPeer: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, peerPubkey)
		if peerPoolItem.TotalPos > uint64(globalParam.PosLimit)*peerPoolItem.InitPos {
			//log.Debugf("voteForPeer: TotalPos: %d : poslimit: %d, InitPos: %d", peerPoolItem.TotalPos, globalParam.PosLimit, peerPoolItem.InitPos)
			return errors.NewErr("voteForPeer, pos of this peer is full!")
		}

		peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
		err = putVoteInfo(native, contract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}

	switch flag {
	case "transfer":
		//zpt transfer
		err = appCallTransferZpt(native, params.Address, utils.GovernanceContractAddress, total)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferZpt, zpt transfer error!")
		}
	case "transferFrom":
		//zpt transfer from
		err = appCallTransferFromZpt(native, utils.GovernanceContractAddress, params.Address, utils.GovernanceContractAddress, total)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromZpt, zpt transfer error!")
		}
	}

	//update total stake
	err = depositTotalStake(native, contract, params.Address, total)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "depositTotalStake, depositTotalStake error!")
	}

	return nil
}

func normalQuit(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	flag := false
	//draw back vote pos
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		voteInfo.WithdrawUnfreezePos = voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos + voteInfo.WithdrawPos +
			voteInfo.WithdrawFreezePos + voteInfo.WithdrawUnfreezePos
		voteInfo.ConsensusPos = 0
		voteInfo.FreezePos = 0
		voteInfo.NewPos = 0
		voteInfo.WithdrawPos = 0
		voteInfo.WithdrawFreezePos = 0
		if voteInfo.Address == peerPoolItem.Address {
			flag = true
			voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + peerPoolItem.InitPos
		}
		err = putVoteInfo(native, contract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	if flag == false {
		voteInfo := &VoteInfo{
			PeerPubkey:          peerPoolItem.PeerPubkey,
			Address:             peerPoolItem.Address,
			WithdrawUnfreezePos: peerPoolItem.InitPos,
		}
		err = putVoteInfo(native, contract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func blackQuit(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	// zpt transfer to trigger unboundGala
	err := appCallTransferZpt(native, utils.GovernanceContractAddress, utils.GovernanceContractAddress, peerPoolItem.InitPos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferZpt, zpt transfer error!")
	}

	//update total stake
	err = withdrawTotalStake(native, contract, peerPoolItem.Address, peerPoolItem.InitPos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "withdrawTotalStake, withdrawTotalStake error!")
	}

	initPos := peerPoolItem.InitPos
	var votePos uint64

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//draw back vote pos
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		total := voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos + voteInfo.WithdrawPos + voteInfo.WithdrawFreezePos
		penalty := (uint64(globalParam.Penalty)*total + 99) / 100
		voteInfo.WithdrawUnfreezePos = total - penalty + voteInfo.WithdrawUnfreezePos
		voteInfo.ConsensusPos = 0
		voteInfo.FreezePos = 0
		voteInfo.NewPos = 0
		voteInfo.WithdrawPos = 0
		voteInfo.WithdrawFreezePos = 0
		address := voteInfo.Address
		err = putVoteInfo(native, contract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}

		//update total stake
		err = withdrawTotalStake(native, contract, address, penalty)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "withdrawTotalStake, withdrawTotalStake error!")
		}
		votePos = votePos + penalty
	}

	//add penalty stake
	err = depositPenaltyStake(native, contract, peerPoolItem.PeerPubkey, initPos, votePos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "depositPenaltyStake, deposit penaltyStake error!")
	}
	return nil
}

func consensusToConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		if voteInfo.FreezePos != 0 {
			return errors.NewErr("commitPos, freezePos should be 0!")
		}
		newPos := voteInfo.NewPos
		voteInfo.ConsensusPos = voteInfo.ConsensusPos + newPos
		voteInfo.NewPos = 0
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		err = putVoteInfo(native, contract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func unConsensusToConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		if voteInfo.ConsensusPos != 0 {
			return errors.NewErr("consensusPos, freezePos should be 0!")
		}

		voteInfo.ConsensusPos = voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos
		voteInfo.NewPos = 0
		voteInfo.FreezePos = 0
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		err = putVoteInfo(native, contract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func consensusToUnConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		if voteInfo.FreezePos != 0 {
			return errors.NewErr("commitPos, freezePos should be 0!")
		}

		voteInfo.FreezePos = voteInfo.ConsensusPos + voteInfo.NewPos
		voteInfo.NewPos = 0
		voteInfo.ConsensusPos = 0
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		err = putVoteInfo(native, contract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func unConsensusToUnConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		if voteInfo.ConsensusPos != 0 {
			return errors.NewErr("consensusPos, freezePos should be 0!")
		}

		newPos := voteInfo.NewPos
		freezePos := voteInfo.FreezePos
		voteInfo.NewPos = 0
		voteInfo.FreezePos = newPos + freezePos
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		err = putVoteInfo(native, contract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func depositTotalStake(native *native.NativeService, contract common.Address, address common.Address, stake uint64) error {
	totalStake, err := getTotalStake(native, contract, address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getTotalStake, get totalStake error!")
	}

	preStake := totalStake.Stake
	preTimeOffset := totalStake.TimeOffset
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP
	//log.Debugf("depositTotalStake: preTimeOffset: %d, timeOffset: %d", preTimeOffset, timeOffset)

	amount := utils.CalcUnbindGala(preStake, preTimeOffset, timeOffset)
	err = appCallTransferFromGala(native, utils.GovernanceContractAddress, utils.ZptContractAddress, totalStake.Address, amount)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromGala, transfer from gala error!")
	}

	totalStake.Stake = preStake + stake
	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, contract, totalStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putTotalStake, put totalStake error!")
	}
	return nil
}

func withdrawTotalStake(native *native.NativeService, contract common.Address, address common.Address, stake uint64) error {
	totalStake, err := getTotalStake(native, contract, address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getTotalStake, get totalStake error!")
	}
	if totalStake.Stake < stake {
		return errors.NewErr("withdraw, zpt deposit is not enough!")
	}

	preStake := totalStake.Stake
	preTimeOffset := totalStake.TimeOffset
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindGala(preStake, preTimeOffset, timeOffset)
	err = appCallTransferFromGala(native, utils.GovernanceContractAddress, utils.ZptContractAddress, totalStake.Address, amount)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromGala, transfer from gala error!")
	}

	totalStake.Stake = preStake - stake
	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, contract, totalStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putTotalStake, put totalStake error!")
	}
	return nil
}

func depositPenaltyStake(native *native.NativeService, contract common.Address, peerPubkey string, initPos uint64, votePos uint64) error {
	penaltyStake, err := getPenaltyStake(native, contract, peerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPenaltyStake, get penaltyStake error!")
	}

	preInitPos := penaltyStake.InitPos
	preVotePos := penaltyStake.VotePos
	preStake := preInitPos + preVotePos
	preTimeOffset := penaltyStake.TimeOffset
	preAmount := penaltyStake.Amount
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindGala(preStake, preTimeOffset, timeOffset)

	penaltyStake.Amount = preAmount + amount
	penaltyStake.InitPos = preInitPos + initPos
	penaltyStake.VotePos = preVotePos + votePos
	penaltyStake.TimeOffset = timeOffset

	err = putPenaltyStake(native, contract, penaltyStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPenaltyStake, put penaltyStake error!")
	}
	return nil
}

func withdrawPenaltyStake(native *native.NativeService, contract common.Address, peerPubkey string, address common.Address) error {
	penaltyStake, err := getPenaltyStake(native, contract, peerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPenaltyStake, get penaltyStake error!")
	}

	preStake := penaltyStake.InitPos + penaltyStake.VotePos
	preTimeOffset := penaltyStake.TimeOffset
	preAmount := penaltyStake.Amount
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindGala(preStake, preTimeOffset, timeOffset)

	//zpt transfer
	err = appCallTransferZpt(native, utils.GovernanceContractAddress, address, preStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferZpt, zpt transfer error!")
	}
	//gala approve
	err = appCallTransferFromGala(native, utils.GovernanceContractAddress, utils.ZptContractAddress, address, amount+preAmount)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromGala, transfer from gala error!")
	}

	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PENALTY_STAKE), peerPubkeyPrefix))
	return nil
}

func executeCommitDpos(native *native.NativeService, contract common.Address, config *Configuration) error {
	//get governace view
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGovernanceView, get GovernanceView error!")
	}

	//get current view
	view := governanceView.View
	newView := view + 1

	//get peerPoolMap
	peerPoolMapSplit, err := GetPeerPoolMap(native, contract, view-1)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//feeSplit first
	log.Debugf("executeCommitDpos executeSplit\n")
	err = executeSplit(native, contract, peerPoolMapSplit)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, executeSplit error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	var peers []*PeerStakeInfo
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == QuitingStatus {
			err = normalQuit(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "normalQuit, normalQuit error!")
			}
			log.Infof("delete normalQuit: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, peerPoolItem.PeerPubkey)
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == BlackStatus {
			err = blackQuit(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "blackQuit, blackQuit error!")
			}
			log.Infof("delete blackQuit: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, peerPoolItem.PeerPubkey)
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == QuitConsensusStatus {
			peerPoolItem.Status = QuitingStatus
			peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
		}
		if peerPoolItem.Status == QuitCandidateStatus {
			err = normalQuit(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "normalQuit, normalQuit error!")
			}
			log.Infof("delete QuitCandidateStatus: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, peerPoolItem.PeerPubkey)
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peers = append(peers, &PeerStakeInfo{
				Index:      peerPoolItem.Index,
				PeerPubkey: peerPoolItem.PeerPubkey,
				Stake:      stake,
			})
		}
	}
	if len(peers) < int(config.K) {
		return errors.NewErr("commitDpos, num of peers is less than K!")
	}

	// sort peers by stake
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].Stake > peers[j].Stake {
			return true
		} else if peers[i].Stake == peers[j].Stake {
			return peers[i].PeerPubkey > peers[j].PeerPubkey
		}
		return false
	})

	// consensus peers
	for i := 0; i < int(config.K); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return errors.NewErr("commitDpos, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "consensusToConsensus, consensusToConsensus error!")
			}
		} else {
			err = unConsensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "unConsensusToConsensus, unConsensusToConsensus error!")
			}
		}
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "consensusToUnConsensus, consensusToUnConsensus error!")
			}
		} else {
			err = unConsensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "unConsensusToUnConsensus, unConsensusToUnConsensus error!")
			}
		}
		peerPoolItem.Status = CandidateStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}
	err = putPeerPoolMap(native, contract, newView, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		log.Infof("comitPos: peerPoolItem: %+v", peerPoolItem)
	}
	oldView := view - 1
	oldViewBytes, err := GetUint32Bytes(oldView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get oldViewBytes error!")
	}
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), oldViewBytes))

	//update view
	governanceView = &GovernanceView{
		View:   newView,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	err = putGovernanceView(native, contract, governanceView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putGovernanceView, put governanceView error!")
	}

	return nil
}

func executeSplit(native *native.NativeService, contract common.Address, peerPoolMap *PeerPoolMap) error {
	balance, err := getGalaBalance(native, utils.GovernanceContractAddress)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, getGalaBalance error!")
	}
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	peersCandidate := []*CandidateSplitInfo{}

	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peersCandidate = append(peersCandidate, &CandidateSplitInfo{
				PeerPubkey: peerPoolItem.PeerPubkey,
				InitPos:    peerPoolItem.InitPos,
				Address:    peerPoolItem.Address,
				Stake:      stake,
			})
		}
	}

	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getConfig, get config error!")
	}

	// sort peers by stake
	sort.SliceStable(peersCandidate, func(i, j int) bool {
		if peersCandidate[i].Stake > peersCandidate[j].Stake {
			return true
		} else if peersCandidate[i].Stake == peersCandidate[j].Stake {
			return peersCandidate[i].PeerPubkey > peersCandidate[j].PeerPubkey
		}
		return false
	})

	// cal s of each consensus node
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	// if sum = 0, means consensus peer in config, do not split
	if sum < uint64(config.K) {
		return nil
	}
	avg := sum / uint64(config.K)
	var sumS uint64
	for i := 0; i < int(config.K); i++ {
		peersCandidate[i].S, err = splitCurve(native, contract, peersCandidate[i].Stake, avg, uint64(globalParam.Yita))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "splitCurve, calculate splitCurve error!")
		}
		sumS += peersCandidate[i].S
	}
	if sumS == 0 {
		return errors.NewErr("executeSplit, sumS is 0!")
	}

	//fee split of consensus peer
	//log.Debugf("fee split of consensus peer")
	for i := int(config.K) - 1; i >= 0; i-- {
		distributedBalance := float64(balance) * float64(globalParam.A) / float64(100)
		proportion := float64(peersCandidate[i].S) / float64(sumS)
		nodeAmount := distributedBalance * proportion
		address := peersCandidate[i].Address
		err = appCallTransferGala(native, utils.GovernanceContractAddress, address, uint64(nodeAmount))
		log.Infof("consensus peer split balance: %d, globalParam.A: %d, peersCandidate[i].S:%d, sumS:%d distributedBalance: %f, proportion: %f, nodeAmount: %+v",
			balance,
			globalParam.A,
			peersCandidate[i].S,
			sumS,
			distributedBalance,
			proportion,
			nodeAmount)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, gala transfer error!")
		}
	}

	//fee split of candidate peer
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	if sum == 0 {
		return nil
	}
	if native.Height >= 720000 {
		for i := int(config.K); i < len(peersCandidate); i++ {
			distributedBalance := float64(balance) * float64(globalParam.B) / float64(100)
			proportion := float64(peersCandidate[i].Stake) / float64(sum)
			nodeAmount := distributedBalance * proportion
			address := peersCandidate[i].Address
			err = appCallTransferGala(native, utils.GovernanceContractAddress, address, uint64(nodeAmount))
			log.Infof("candidate peer split balance: %d, globalParam.B: %d, peersCandidate[i].Stake:%d, sum:%d distributedBalance: %f, proportion: %f, nodeAmount: %+v",
				balance,
				globalParam.B,
				peersCandidate[i].Stake,
				sum,
				distributedBalance,
				proportion,
				nodeAmount)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, gala transfer error!")
			}
		}
	} else {
		for i := int(config.K); i < len(peersCandidate); i++ {
			distributedBalance := float64(balance) * float64(globalParam.B) / float64(100)
			proportion := float64(peersCandidate[i].S) / float64(sumS)
			nodeAmount := distributedBalance * proportion
			address := peersCandidate[i].Address
			err = appCallTransferGala(native, utils.GovernanceContractAddress, address, uint64(nodeAmount))
			log.Infof("candidate peer split balance: %d, globalParam.B: %d, peersCandidate[i].Stake:%d, sum:%d distributedBalance: %f, proportion: %f, nodeAmount: %+v",
				balance,
				globalParam.B,
				peersCandidate[i].Stake,
				sum,
				distributedBalance,
				proportion,
				nodeAmount)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, gala transfer error!")
			}
		}
	}

	return nil
}
