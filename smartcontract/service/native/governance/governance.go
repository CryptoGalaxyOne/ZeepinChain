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

//Governance contract:
//Users can apply for a candidate node to join consensus selection, deposit ZPT to vote for candidate nodes, quit selection and unVote for candidate nodes through this contract.
//ZPT deposited in the contract can get GALA bonus which come from transaction fee of the network.
package governance

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/imZhuFei/zeepin/common"
	"github.com/imZhuFei/zeepin/common/config"
	"github.com/imZhuFei/zeepin/common/constants"
	"github.com/imZhuFei/zeepin/common/log"
	"github.com/imZhuFei/zeepin/common/serialization"
	cstates "github.com/imZhuFei/zeepin/core/states"
	scommon "github.com/imZhuFei/zeepin/core/store/common"
	"github.com/imZhuFei/zeepin/errors"
	"github.com/imZhuFei/zeepin/smartcontract/service/native"
	"github.com/imZhuFei/zeepin/smartcontract/service/native/global_params"
	"github.com/imZhuFei/zeepin/smartcontract/service/native/utils"
)

const (
	//status
	RegisterCandidateStatus Status = iota
	CandidateStatus
	ConsensusStatus
	QuitConsensusStatus
	QuitingStatus
	QuitCandidateStatus
	BlackStatus
)

const (
	//function name
	INIT_CONFIG                      = "initConfig"
	REGISTER_CANDIDATE               = "registerCandidate"
	REGISTER_CANDIDATE_TRANSFER_FROM = "registerCandidateTransferFrom"
	UNREGISTER_CANDIDATE             = "unRegisterCandidate"
	APPROVE_CANDIDATE                = "approveCandidate"
	REJECT_CANDIDATE                 = "rejectCandidate"
	BLACK_NODE                       = "blackNode"
	WHITE_NODE                       = "whiteNode"
	QUIT_NODE                        = "quitNode"
	VOTE_FOR_PEER                    = "voteForPeer"
	VOTE_FOR_PEER_TRANSFER_FROM      = "voteForPeerTransferFrom"
	UNVOTE_FOR_PEER                  = "unVoteForPeer"
	WITHDRAW                         = "withdraw"
	COMMIT_DPOS                      = "commitDpos"
	UPDATE_CONFIG                    = "updateConfig"
	UPDATE_GLOBAL_PARAM              = "updateGlobalParam"
	UPDATE_SPLIT_CURVE               = "updateSplitCurve"
	CALL_SPLIT                       = "callSplit"
	TRANSFER_PENALTY                 = "transferPenalty"
	WITHDRAW_GALA                    = "withdrawGala"
	GET_PEERPOOL_INFO                = "getPeerPoolInfo"
	GET_VOTE_INFO                    = "getVoteInfo"
	CHECK_VOTE_INFO                  = "checkVoteInfo"
	//key prefix
	GLOBAL_PARAM    = "globalParam"
	VBFT_CONFIG     = "vbftConfig"
	GOVERNANCE_VIEW = "governanceView"
	CANDIDITE_INDEX = "candidateIndex"
	PEER_POOL       = "peerPool"
	VOTE_INFO_POOL  = "voteInfoPool"
	PEER_INDEX      = "peerIndex"
	BLACK_LIST      = "blackList"
	TOTAL_STAKE     = "totalStake"
	PENALTY_STAKE   = "penaltyStake"
	SPLIT_CURVE     = "splitCurve"

	//global
	PRECISE = 1000000
)

// candidate fee must >= 1 Gala
var MinCandidateFee = uint64(math.Pow(10, constants.GALA_DECIMALS))

var Xi = []uint32{
	0, 100000, 200000, 300000, 400000, 500000, 600000, 700000, 800000, 900000, 1000000, 1100000, 1200000, 1300000, 1400000,
	1500000, 1600000, 1700000, 1800000, 1900000, 2000000, 2100000, 2200000, 2300000, 2400000, 2500000, 2600000, 2700000,
	2800000, 2900000, 3000000, 3100000, 3200000, 3300000, 3400000, 3500000, 3600000, 3700000, 3800000, 3900000, 4000000,
	4100000, 4200000, 4300000, 4400000, 4500000, 4600000, 4700000, 4800000, 4900000, 5000000, 5100000, 5200000, 5300000,
	5400000, 5500000, 5600000, 5700000, 5800000, 5900000, 6000000, 6100000, 6200000, 6300000, 6400000, 6500000, 6600000,
	6700000, 6800000, 6900000, 7000000, 7100000, 7200000, 7300000, 7400000, 7500000, 7600000, 7700000, 7800000, 7900000,
	8000000, 8100000, 8200000, 8300000, 8400000, 8500000, 8600000, 8700000, 8800000, 8900000, 9000000, 9100000, 9200000,
	9300000, 9400000, 9500000, 9600000, 9700000, 9800000, 9900000, 10000000,
}

//Init governance contract address
func InitGovernance() {
	native.Contracts[utils.GovernanceContractAddress] = RegisterGovernanceContract
}

//Register methods of governance contract
func RegisterGovernanceContract(native *native.NativeService) {
	native.Register(REGISTER_CANDIDATE, RegisterCandidate)
	native.Register(REGISTER_CANDIDATE_TRANSFER_FROM, RegisterCandidateTransferFrom)
	native.Register(UNREGISTER_CANDIDATE, UnRegisterCandidate)
	native.Register(VOTE_FOR_PEER, VoteForPeer)
	native.Register(VOTE_FOR_PEER_TRANSFER_FROM, VoteForPeerTransferFrom)
	native.Register(UNVOTE_FOR_PEER, UnVoteForPeer)
	native.Register(WITHDRAW, Withdraw)
	native.Register(QUIT_NODE, QuitNode)
	native.Register(WITHDRAW_GALA, WithdrawGala)

	native.Register(INIT_CONFIG, InitConfig)
	native.Register(APPROVE_CANDIDATE, ApproveCandidate)
	native.Register(REJECT_CANDIDATE, RejectCandidate)
	native.Register(BLACK_NODE, BlackNode)
	native.Register(WHITE_NODE, WhiteNode)
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(UPDATE_CONFIG, UpdateConfig)
	native.Register(UPDATE_GLOBAL_PARAM, UpdateGlobalParam)
	native.Register(UPDATE_SPLIT_CURVE, UpdateSplitCurve)
	native.Register(CALL_SPLIT, CallSplit)
	native.Register(TRANSFER_PENALTY, TransferPenalty)
	native.Register(GET_PEERPOOL_INFO, GetPeerpoolInfo)
	native.Register(GET_VOTE_INFO, GetVoteInfo)
	native.Register(CHECK_VOTE_INFO, CheckVoteInfo)
}

//Init governance contract, include vbft config, global param and Gid admin.
func InitConfig(native *native.NativeService) ([]byte, error) {
	configuration := new(config.VBFTConfig)
	buf, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarBytes, contract params deserialize error!")
	}
	if err := configuration.Deserialize(bytes.NewBuffer(buf)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	// check if initConfig is already execute
	governanceViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getGovernanceView, get governanceViewBytes error!")
	}
	if governanceViewBytes != nil {
		return utils.BYTE_FALSE, errors.NewErr("initConfig. initConfig is already executed!")
	}

	//check the configuration
	err = CheckVBFTConfig(configuration)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "checkGBFTConfig failed!")
	}

	//init globalParam
	globalParam := &GlobalParam{
		CandidateFee: 100000000,
		MinInitStake: configuration.MinInitStake,
		CandidateNum: 51,
		PosLimit:     20,
		A:            40,
		B:            60,
		Yita:         5,
		Penalty:      5,
	}
	err = putGlobalParam(native, contract, globalParam)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putGlobalParam, put globalParam error!")
	}

	var view uint32 = 1
	var maxId uint32

	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	for _, peer := range configuration.Peers {
		if peer.Index > maxId {
			maxId = peer.Index
		}
		address, err := common.AddressFromBase58(peer.Address)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
		}

		peerPoolItem := new(PeerPoolItem)
		peerPoolItem.Index = peer.Index
		peerPoolItem.PeerPubkey = peer.PeerPubkey
		peerPoolItem.Address = address
		peerPoolItem.InitPos = peer.InitPos
		peerPoolItem.TotalPos = 0
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem

		peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}
		index := peerPoolItem.Index
		indexBytes, err := GetUint32Bytes(index)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getUint32Bytes, getUint32Bytes error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: indexBytes})

		//update total stake
		err = depositTotalStake(native, contract, address, peerPoolItem.InitPos)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "depositTotalStake, depositTotalStake error!")
		}
	}

	//init peer pool
	err = putPeerPoolMap(native, contract, 0, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}
	indexBytes, err := GetUint32Bytes(maxId + 1)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getUint32Bytes, get indexBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)), &cstates.StorageItem{Value: indexBytes})

	//init governance view
	governanceView := &GovernanceView{
		View:   view,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	err = putGovernanceView(native, contract, governanceView)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putGovernanceView, put governanceView error!")
	}

	//init config
	config := &Configuration{
		N:                    configuration.N,
		C:                    configuration.C,
		K:                    configuration.K,
		L:                    configuration.L,
		BlockMsgDelay:        configuration.BlockMsgDelay,
		HashMsgDelay:         configuration.HashMsgDelay,
		PeerHandshakeTimeout: configuration.PeerHandshakeTimeout,
		MaxBlockChangeView:   configuration.MaxBlockChangeView,
	}
	err = putConfig(native, contract, config)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putConfig, put config error!")
	}

	//init splitCurve
	splitCurve := &SplitCurve{
		Yi: []uint32{
			0, 95123, 180968, 258213, 327493, 389401, 444491, 493282, 536257, 573866, 606531, 634645, 658574, 678660, 695220, 708550,
			718927, 726606, 731826, 734808, 735759, 734870, 732317, 728265, 722867, 716262, 708583, 699949, 690472, 680254, 669391,
			657969, 646069, 633765, 621124, 608209, 595076, 581778, 568361, 554869, 541342, 527814, 514317, 500882, 487534, 474297,
			461191, 448236, 435447, 422839, 410425, 398217, 386223, 374452, 362910, 351604, 340537, 329713, 319135, 308805, 298723,
			288890, 279306, 269969, 260879, 252033, 243429, 235066, 226939, 219045, 211382, 203945, 196731, 189736, 182955, 176384,
			170018, 163854, 157887, 152113, 146526, 141122, 135896, 130845, 125963, 121246, 116690, 112290, 108041, 103940, 99981,
			96162, 92477, 88923, 85496, 82192, 79006, 75936, 72977, 70126, 67380,
		},
	}
	err = putSplitCurve(native, contract, splitCurve)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putSplitCurve, put splitCurve error!")
	}

	//init admin GID
	err = appCallInitContractAdmin(native, []byte(configuration.AdminGID))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallInitContractAdmin error!")
	}

	return utils.BYTE_TRUE, nil
}

//Get peerpool info, used by users.
func GetPeerpoolInfo(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	return bf.Bytes(), nil
}

func GetVoteInfo(native *native.NativeService) ([]byte, error) {
	params := new(GetVoteInfoParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	voteInfo, err := getVoteInfo(native, contract, params.PeerPubkey, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getVoteInfo, get voteInfo error!")
	}
	bf := new(bytes.Buffer)
	if err := voteInfo.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	return bf.Bytes(), nil
}

//Register a candidate node, used by users.
//Users can register a candidate node with a authorized Gid.
//Candidate node can be voted and become consensus node according to their pos.
//Candidate node can get gala bonus according to their pos.
func RegisterCandidate(native *native.NativeService) ([]byte, error) {
	err := registerCandidate(native, "transfer")
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "registerCandidate error!")
	}
	return utils.BYTE_TRUE, nil
}

//Register a candidate node, used by contracts.
//Contracts can register a candidate node with a authorized Gid after approving ZPT to governance contract before invoke this function.
//Candidate node can be voted and become consensus node according to their pos.
//Candidate node can get gala bonus according to their pos.
func RegisterCandidateTransferFrom(native *native.NativeService) ([]byte, error) {
	err := registerCandidate(native, "transferFrom")
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "registerCandidateTransferFrom error!")
	}
	return utils.BYTE_TRUE, nil
}

//Unregister a registered candidate node, will remove node from pool, and unfreeze deposit zpt.
func UnRegisterCandidate(native *native.NativeService) ([]byte, error) {
	params := new(UnRegisterCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address := params.Address
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//check if exist in PeerPool
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("unRegisterCandidate, peerPubkey is not in peerPoolMap!")
	}

	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, errors.NewErr("unRegisterCandidate, peer status is not RegisterCandidateStatus!")
	}

	//check owner address
	if peerPoolItem.Address != params.Address {
		return utils.BYTE_FALSE, errors.NewErr("unRegisterCandidate, address is not peer owner!")
	}

	//unfreeze initPos
	voteInfo := &VoteInfo{
		PeerPubkey:          peerPoolItem.PeerPubkey,
		Address:             peerPoolItem.Address,
		WithdrawUnfreezePos: 0,
	}
	err = putVoteInfo(native, contract, voteInfo)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
	}

	//change peerPool status

	peerPoolItem.Status = QuitCandidateStatus

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	log.Infof("UnregisterCandidate: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, peerPoolItem.PeerPubkey)

	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}
	return utils.BYTE_TRUE, nil
}

//Approve a registered candidate node, used by admin.
//Only approved candidate node can participate in consensus selection and get gala bonus.
func ApproveCandidate(native *native.NativeService) ([]byte, error) {
	params := new(ApproveCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "approveCandidate, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	//check if peerPoolMap full
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	num := 0
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			num = num + 1
		}
	}
	if num >= int(globalParam.CandidateNum) {
		return utils.BYTE_FALSE, errors.NewErr("approveCandidate, num of candidate node is full!")
	}

	//get peerPool
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("approveCandidate, peerPubkey is not in peerPoolMap!")
	}

	//check initPos
	if peerPoolItem.InitPos < uint64(globalParam.MinInitStake) {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, initPos %v must >= %v", peerPoolItem.InitPos, globalParam.MinInitStake)
	}

	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, errors.NewErr("approveCandidate, peer status is not RegisterCandidateStatus!")
	}
	log.Infof("ApproveCandidate: status: %d : peerPubkey: %s", peerPoolItem.Status, params.PeerPubkey)
	peerPoolItem.Status = CandidateStatus
	if native.Height < 619000 {
		peerPoolItem.TotalPos = 0
	}

	log.Infof("ApproveCandidate: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, params.PeerPubkey)
	//check if has index
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	indexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get indexBytes error!")
	}
	if indexBytes != nil {
		index, err := GetBytesUint32(indexBytes.(*cstates.StorageItem).Value)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetBytesUint32, get index error!")
		}
		peerPoolItem.Index = index
	} else {
		//get candidate index
		candidateIndex, err := getCandidateIndex(native, contract)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getCandidateIndex, get candidateIndex error!")
		}
		peerPoolItem.Index = candidateIndex

		//update candidateIndex
		newCandidateIndex := candidateIndex + 1
		err = putCandidateIndex(native, contract, newCandidateIndex)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "putCandidateIndex, put candidateIndex error!")
		}

		indexBytes, err := GetUint32Bytes(peerPoolItem.Index)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get indexBytes error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: indexBytes})
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}

	return utils.BYTE_TRUE, nil
}

//Reject a registered candidate node, remove node from pool and unfreeze deposit zpt, used by admin.
//Only approved candidate node can participate in consensus selection and get gala bonus.
func RejectCandidate(native *native.NativeService) ([]byte, error) {
	params := new(RejectCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "approveCandidate, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//draw back init pos
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("rejectCandidate, peerPubkey is not in peerPoolMap!")
	}
	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, errors.NewErr("rejectCandidate, peerPubkey is not RegisterCandidateStatus!")
	}
	address := peerPoolItem.Address
	voteInfo, err := getVoteInfo(native, contract, params.PeerPubkey, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getVoteInfo, get voteInfo error!")
	}
	//voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + peerPoolItem.InitPos
	err = putVoteInfo(native, contract, voteInfo)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
	}

	//remove peerPubkey from peerPool
	peerPoolItem.Status = QuitCandidateStatus

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	log.Infof("rejectNode: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, peerPoolItem.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}

	return utils.BYTE_TRUE, nil
}

//Put a node into black list, remove node from pool, used by admin.
//Whole of initPos of black node will be punished, and several percent of vote deposit will be punished too.
//Node in black list can't be registered.
func BlackNode(native *native.NativeService) ([]byte, error) {
	params := new(BlackNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "blackNode, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}
	commit := false
	for _, peerPubkey := range params.PeerPubkeyList {
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("blackNode, peerPubkey is not in peerPoolMap!")
		}

		blackListItem := &BlackListItem{
			PeerPubkey: peerPoolItem.PeerPubkey,
			Address:    peerPoolItem.Address,
			InitPos:    peerPoolItem.InitPos,
		}
		bf := new(bytes.Buffer)
		if err := blackListItem.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize blackListItem error!")
		}
		//put peer into black list
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix), &cstates.StorageItem{Value: bf.Bytes()})
		//change peerPool status
		if peerPoolItem.Status == ConsensusStatus {
			peerPoolItem.Status = BlackStatus
			peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
			err = putPeerPoolMap(native, contract, view, peerPoolMap)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
			}
			commit = true
		} else {
			peerPoolItem.Status = BlackStatus
			peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
			err = putPeerPoolMap(native, contract, view, peerPoolMap)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
			}
		}
	}
	//commitDpos
	if commit {
		// get config
		config, err := getConfig(native, contract)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getConfig, get config error!")
		}
		err = executeCommitDpos(native, contract, config)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeCommitDpos, executeCommitDpos error!")
		}
	}
	return utils.BYTE_TRUE, nil
}

//Remove a node from black list, allow it to be registered, used by admin.
func WhiteNode(native *native.NativeService) ([]byte, error) {
	params := new(WhiteNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "whiteNode, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}

	//check black list
	blackListBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get BlackList error!")
	}
	if blackListBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("whiteNode, this Peer is not in BlackList!")
	}

	//remove peer from black list
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))

	return utils.BYTE_TRUE, nil
}

//Quit a registered node, used by node owner.
//Remove node from pool and unfreeze deposit next epoch(candidate node) / next next epoch(consensus node)
func QuitNode(native *native.NativeService) ([]byte, error) {
	params := new(QuitNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//get config
	config, err := getConfig(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getConfig, get config error!")
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("quitNode, peerPubkey is not in peerPoolMap!")
	}

	if address != peerPoolItem.Address {
		return utils.BYTE_FALSE, errors.NewErr("quitNode, peerPubkey is not registered by this address!")
	}
	if peerPoolItem.Status != ConsensusStatus && peerPoolItem.Status != CandidateStatus {
		return utils.BYTE_FALSE, errors.NewErr("quitNode, peerPubkey is not CandidateStatus or ConsensusStatus!")
	}

	//check peers num
	num := 0
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			num = num + 1
		}
	}
	if num <= int(config.K) {
		return utils.BYTE_FALSE, errors.NewErr("quitNode, num of peers is less than K!")
	}

	//change peerPool status
	if peerPoolItem.Status == ConsensusStatus {
		peerPoolItem.Status = QuitConsensusStatus
	} else {
		peerPoolItem.Status = QuitingStatus
	}

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	log.Infof("QuitNode: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, params.PeerPubkey)
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}

	return utils.BYTE_TRUE, nil
}

//Vote for a node by depositing ZPT in this governance contract, used by users
func VoteForPeer(native *native.NativeService) ([]byte, error) {
	err := voteForPeer(native, "transfer")
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "voteForPeer error!")
	}
	return utils.BYTE_TRUE, nil
}

//Vote for a node by depositing ZPT in this governance contract, used by contracts
func VoteForPeerTransferFrom(native *native.NativeService) ([]byte, error) {
	err := voteForPeer(native, "transferFrom")
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "voteForPeerTransferFrom error!")
	}
	return utils.BYTE_TRUE, nil
}

//UnVote for a node by redeeming ZPT from this governance contract
func UnVoteForPeer(native *native.NativeService) ([]byte, error) {
	params := &VoteForPeerParam{
		PeerPubkeyList: make([]string, 0),
		PosList:        make([]uint64, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := params.PosList[i]

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("unVoteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus && peerPoolItem.Status != RegisterCandidateStatus {
			return utils.BYTE_FALSE, errors.NewErr("unVoteForPeer, peerPubkey is not candidate and can not be unvoted!")
		}

		voteInfo, err := getVoteInfo(native, contract, peerPubkey, address)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getVoteInfo, get voteInfo error!")
		}
		if voteInfo.NewPos < uint64(pos) {
			if peerPoolItem.Status == ConsensusStatus {
				if voteInfo.ConsensusPos < (uint64(pos) - voteInfo.NewPos) {
					return utils.BYTE_FALSE, errors.NewErr("unVoteForPeer, your pos of this peerPubkey is not enough!")
				}
				consensusPos := voteInfo.ConsensusPos + voteInfo.NewPos - uint64(pos)
				newPos := voteInfo.NewPos
				voteInfo.NewPos = 0
				voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + newPos
				voteInfo.ConsensusPos = consensusPos
				voteInfo.WithdrawPos = voteInfo.WithdrawPos + uint64(pos) - voteInfo.NewPos
				peerPoolItem.TotalPos = peerPoolItem.TotalPos - uint64(pos)
			}
			if peerPoolItem.Status == CandidateStatus {
				if voteInfo.FreezePos < (uint64(pos) - voteInfo.NewPos) {
					return utils.BYTE_FALSE, errors.NewErr("unVoteForPeer, your pos of this peerPubkey is not enough!")
				}
				freezePos := voteInfo.FreezePos + voteInfo.NewPos - uint64(pos)
				newPos := voteInfo.NewPos
				voteInfo.NewPos = 0
				voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + newPos
				voteInfo.FreezePos = freezePos
				voteInfo.WithdrawFreezePos = voteInfo.WithdrawFreezePos + uint64(pos) - voteInfo.NewPos
				peerPoolItem.TotalPos = peerPoolItem.TotalPos - uint64(pos)
			}
			if peerPoolItem.Status == RegisterCandidateStatus {
				if voteInfo.FreezePos < (uint64(pos) - voteInfo.NewPos) {
					return utils.BYTE_FALSE, errors.NewErr("unVoteForPeer, your pos of this peerPubkey is not enough!")
				}
				freezePos := voteInfo.FreezePos + voteInfo.NewPos - uint64(pos)
				newPos := voteInfo.NewPos
				voteInfo.NewPos = 0
				voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + newPos
				voteInfo.FreezePos = freezePos
				voteInfo.WithdrawFreezePos = voteInfo.WithdrawFreezePos + uint64(pos) - voteInfo.NewPos
				peerPoolItem.TotalPos = peerPoolItem.TotalPos - uint64(pos)
			}
		} else {
			temp := voteInfo.NewPos - uint64(pos)
			voteInfo.NewPos = temp
			voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + uint64(pos)
			peerPoolItem.TotalPos = peerPoolItem.TotalPos - uint64(pos)
		}

		log.Infof("unvoteForPeer: TotalPos: %d : peerPubkey: %s unvotepos:%d height:%d", peerPoolItem.TotalPos, peerPubkey, uint64(pos), native.Height)
		peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
		err = putVoteInfo(native, contract, voteInfo)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}

	return utils.BYTE_TRUE, nil
}

func CheckVoteInfo(native *native.NativeService) ([]byte, error) {

	params := new(CheckVoteInfoParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}
	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "CheckVoteInfo, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("CheckVoteInfo, peerPubkey is not in peerPoolMap!")
	}
	if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus && peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, errors.NewErr("CheckVoteInfo, peerPubkey is not candidate and can not be Check!")
	}
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	totalPos := uint64(0)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		totalPos += voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos
		log.Info("address:", voteInfo.Address.ToBase58(), "ConsensusPos:", voteInfo.ConsensusPos, "FreezePos:", voteInfo.FreezePos, "NewPos:", voteInfo.NewPos, "pubKey:", voteInfo.PeerPubkey)
		log.Info("address:", voteInfo.Address.ToBase58(), voteInfo, voteInfo.PeerPubkey)

	}
	peerPoolItem.TotalPos = totalPos
	if peerPoolItem.TotalPos > uint64(globalParam.PosLimit)*peerPoolItem.InitPos {
		//log.Debugf("voteForPeer: TotalPos: %d : poslimit: %d, InitPos: %d", peerPoolItem.TotalPos, globalParam.PosLimit, peerPoolItem.InitPos)
		return utils.BYTE_FALSE, errors.NewErr("CheckVoteInfo, pos of this peer is full!")
	}
	if peerPoolItem.TotalPos < 0 {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "total pos must greater than zero!")
	}
	log.Infof("CheckVoteInfo: TotalPos: %d : peerPubkey: %s", peerPoolItem.TotalPos, peerPoolItem.PeerPubkey)
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}
	return utils.BYTE_TRUE, nil
}

//Withdraw unfreezed ZPT deposited in this governance contract.
func Withdraw(native *native.NativeService) ([]byte, error) {
	params := &WithdrawParam{
		PeerPubkeyList: make([]string, 0),
		WithdrawList:   make([]uint64, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	var total uint64
	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := params.WithdrawList[i]
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		voteInfo, err := getVoteInfo(native, contract, peerPubkey, address)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getVoteInfo, get voteInfo error!")
		}
		if voteInfo.WithdrawUnfreezePos < uint64(pos) {
			return utils.BYTE_FALSE, errors.NewErr("withdraw, your unfreeze withdraw pos of this peerPubkey is not enough!")
		} else {
			voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos - uint64(pos)
			total = total + uint64(pos)
			err = putVoteInfo(native, contract, voteInfo)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
			}
		}
		if voteInfo.ConsensusPos == 0 && voteInfo.FreezePos == 0 && voteInfo.NewPos == 0 &&
			voteInfo.WithdrawPos == 0 && voteInfo.WithdrawFreezePos == 0 && voteInfo.WithdrawUnfreezePos == 0 {
			native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix, address[:]))
		}
	}

	//ZPT transfer
	err = appCallTransferZpt(native, utils.GovernanceContractAddress, address, total)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferZpt, ZPT transfer error!")
	}

	//update total stake
	err = withdrawTotalStake(native, contract, address, total)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "withdrawTotalStake, withdrawTotalStake error!")
	}

	return utils.BYTE_TRUE, nil
}

//Go to next consensus epoch
func CommitDpos(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getConfig, get config error!")
	}

	//get governace view
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getGovernanceView, get GovernanceView error!")
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		cycle := (native.Height - governanceView.Height) >= config.MaxBlockChangeView
		if !cycle {
			return utils.BYTE_FALSE, errors.NewErr("commitDpos, authentication Failed!")
		}
	}
	err = executeCommitDpos(native, contract, config)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeCommitDpos, executeCommitDpos error!")
	}

	return utils.BYTE_TRUE, nil
}

//Update VBFT config
func UpdateConfig(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "updateConfig, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	configuration := new(Configuration)
	if err := configuration.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize configuration error!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}
	candidateNum := 0
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			candidateNum = candidateNum + 1
		}
	}

	//check the configuration
	if configuration.C == 0 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. C can not be 0 in config!")
	}
	if int(configuration.K) > candidateNum {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. K can not be larger than num of candidate peer in config!")
	}
	if configuration.L < 16*configuration.K || configuration.L%configuration.K != 0 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. L can not be less than 16*K and K must be times of L in config!")
	}
	if configuration.K < 2*configuration.C+1 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. K can not be less than 2*C+1 in config!")
	}
	if 4*configuration.K > globalParam.CandidateNum {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. 4*K can not be more than candidateNum!")
	}
	if configuration.N < configuration.K || configuration.K < 9 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. config not match N >= K >= 9!")
	}
	if configuration.BlockMsgDelay < 5000 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. BlockMsgDelay must >= 5000!")
	}
	if configuration.HashMsgDelay < 5000 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. HashMsgDelay must >= 5000!")
	}
	if configuration.PeerHandshakeTimeout < 10 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. PeerHandshakeTimeout must >= 10!")
	}
	err = putConfig(native, contract, configuration)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putConfig, put config error!")
	}

	return utils.BYTE_TRUE, nil
}

//Update global params of this governance contract
func UpdateGlobalParam(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "updateGlobalParam, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getConfig, get config error!")
	}

	globalParam := new(GlobalParam)
	if err := globalParam.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize globalParam error!")
	}

	//check the globalParam
	if (globalParam.A + globalParam.B) != 100 {
		return utils.BYTE_FALSE, errors.NewErr("updateGlobalParam. A + B must equal to 100!")
	}
	if globalParam.Yita == 0 {
		return utils.BYTE_FALSE, errors.NewErr("updateGlobalParam. Yita must > 0!")
	}
	if globalParam.Penalty > 100 {
		return utils.BYTE_FALSE, errors.NewErr("updateGlobalParam. Penalty must <= 100!")
	}
	if globalParam.PosLimit < 1 {
		return utils.BYTE_FALSE, errors.NewErr("updateGlobalParam. PosLimit must >= 1!")
	}
	if globalParam.CandidateNum < 4*config.K {
		return utils.BYTE_FALSE, errors.NewErr("updateGlobalParam. CandidateNum must >= 4*K!")
	}
	if globalParam.CandidateFee != 0 && globalParam.CandidateFee < MinCandidateFee {
		return utils.BYTE_FALSE, fmt.Errorf("updateGlobalParam. CandidateFee must >= %d", MinCandidateFee)
	}
	err = putGlobalParam(native, contract, globalParam)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putGlobalParam, put globalParam error!")
	}

	return utils.BYTE_TRUE, nil
}

//Update split curve
func UpdateSplitCurve(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "updateSplitCurve, checkWitness error!")
	}

	splitCurve := new(SplitCurve)
	if err := splitCurve.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize splitCurve error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	err = putSplitCurve(native, contract, splitCurve)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putSplitCurve, put splitCurve error!")
	}

	return utils.BYTE_TRUE, nil
}

//Trigger fee split
func CallSplit(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "callSplit, checkWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	//get current view
	cView, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "callSplit, get view error!")
	}
	view := cView - 1

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "callSplit, get peerPoolMap error!")
	}

	err = executeSplit(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, executeSplitp error!")
	}

	return utils.BYTE_TRUE, nil
}

//Transfer all punished ZPT of a black node to a certain address
func TransferPenalty(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getAdmin, get admin error!")
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "transferPenalty, checkWitness error!")
	}

	param := new(TransferPenaltyParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize transferPenaltyParam error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	err = withdrawPenaltyStake(native, contract, param.PeerPubkey, param.Address)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "withdrawPenaltyStake, withdraw penaltyStake error!")
	}

	return utils.BYTE_TRUE, nil
}

//Withdraw unbounded GALA according to deposit ZPT in this governance contract
func WithdrawGala(native *native.NativeService) ([]byte, error) {
	param := new(WithdrawGalaParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize transferPenaltyParam error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check witness
	err := utils.ValidateOwner(native, param.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "withdrawGala, checkWitness error!")
	}

	// ZPT transfer to trigger unboundGala
	err = appCallTransferZpt(native, utils.GovernanceContractAddress, utils.GovernanceContractAddress, 1)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferZpt, ZPT transfer error!")
	}

	totalStake, err := getTotalStake(native, contract, param.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getTotalStake, get totalStake error!")
	}

	preTimeOffset := totalStake.TimeOffset
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindGala(totalStake.Stake, preTimeOffset, timeOffset)
	err = appCallTransferFromGala(native, utils.GovernanceContractAddress, utils.ZptContractAddress, totalStake.Address, amount)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromGala, transfer from Gala error!")
	}

	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, contract, totalStake)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "putTotalStake, put totalStake error!")
	}
	return utils.BYTE_TRUE, nil
}
