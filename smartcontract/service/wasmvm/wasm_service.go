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
package wasmvm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/imZhuFei/zeepin/common"
	"github.com/imZhuFei/zeepin/core/store"
	"github.com/imZhuFei/zeepin/core/types"
	"github.com/imZhuFei/zeepin/embed/simulator"
	"github.com/imZhuFei/zeepin/errors"
	"github.com/imZhuFei/zeepin/smartcontract/context"
	"github.com/imZhuFei/zeepin/smartcontract/event"
	nstates "github.com/imZhuFei/zeepin/smartcontract/service/native/zpt"
	"github.com/imZhuFei/zeepin/smartcontract/states"
	"github.com/imZhuFei/zeepin/smartcontract/storage"
	"github.com/imZhuFei/zeepin/vm/wasmvm/exec"
	"github.com/imZhuFei/zeepin/vm/wasmvm/util"
)

type WasmVmService struct {
	Store         store.LedgerStore
	CloneCache    *storage.CloneCache
	ContextRef    context.ContextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	Tx            *types.Transaction
	Time          uint32
	Height        uint32
}

func (this *WasmVmService) Invoke() (interface{}, error) {
	stateMachine := NewWasmStateMachine()
	//register the "CallContract" function
	//stateMachine.Register("ZPT_CallContract", this.callContract)
	stateMachine.Register("ZPT_MarshalNativeParams", this.marshalNativeParams)
	stateMachine.Register("ZPT_MarshalEmbededParams", this.marshalEmbeddedParams)
	//runtime
	stateMachine.Register("ZPT_Runtime_CheckWitness", this.runtimeCheckWitness)
	stateMachine.Register("ZPT_Runtime_Notify", this.runtimeNotify)
	stateMachine.Register("ZPT_Runtime_CheckSig", this.runtimeCheckSig)
	stateMachine.Register("ZPT_Runtime_GetTime", this.runtimeGetTime)
	stateMachine.Register("ZPT_Runtime_Log", this.runtimeLog)
	//attribute
	stateMachine.Register("ZPT_Attribute_GetUsage", this.attributeGetUsage)
	stateMachine.Register("ZPT_Attribute_GetData", this.attributeGetData)
	//block
	stateMachine.Register("ZPT_Block_GetCurrentHeaderHash", this.blockGetCurrentHeaderHash)
	stateMachine.Register("ZPT_Block_GetCurrentHeaderHeight", this.blockGetCurrentHeaderHeight)
	stateMachine.Register("ZPT_Block_GetCurrentBlockHash", this.blockGetCurrentBlockHash)
	stateMachine.Register("ZPT_Block_GetCurrentBlockHeight", this.blockGetCurrentBlockHeight)
	stateMachine.Register("ZPT_Block_GetTransactionByHash", this.blockGetTransactionByHash)
	stateMachine.Register("ZPT_Block_GetTransactionCount", this.blockGetTransactionCount)
	stateMachine.Register("ZPT_Block_GetTransactions", this.blockGetTransactions)

	//blockchain
	stateMachine.Register("ZPT_BlockChain_GetHeight", this.blockChainGetHeight)
	stateMachine.Register("ZPT_BlockChain_GetHeaderByHeight", this.blockChainGetHeaderByHeight)
	stateMachine.Register("ZPT_BlockChain_GetHeaderByHash", this.blockChainGetHeaderByHash)
	stateMachine.Register("ZPT_BlockChain_GetBlockByHeight", this.blockChainGetBlockByHeight)
	stateMachine.Register("ZPT_BlockChain_GetBlockByHash", this.blockChainGetBlockByHash)
	stateMachine.Register("ZPT_BlockChain_GetContract", this.blockChainGetContract)

	//header
	stateMachine.Register("ZPT_Header_GetHash", this.headerGetHash)
	stateMachine.Register("ZPT_Header_GetVersion", this.headerGetVersion)
	stateMachine.Register("ZPT_Header_GetPrevHash", this.headerGetPrevHash)
	stateMachine.Register("ZPT_Header_GetMerkleRoot", this.headerGetMerkleRoot)
	stateMachine.Register("ZPT_Header_GetIndex", this.headerGetIndex)
	stateMachine.Register("ZPT_Header_GetTimestamp", this.headerGetTimestamp)
	stateMachine.Register("ZPT_Header_GetConsensusData", this.headerGetConsensusData)
	stateMachine.Register("ZPT_Header_GetNextConsensus", this.headerGetNextConsensus)

	//storage
	stateMachine.Register("ZPT_Storage_Put", this.putstore)
	stateMachine.Register("ZPT_Storage_Get", this.getstore)
	stateMachine.Register("ZPT_Storage_Delete", this.deletestore)

	//transaction
	stateMachine.Register("ZPT_Transaction_GetHash", this.transactionGetHash)
	stateMachine.Register("ZPT_Transaction_GetType", this.transactionGetType)
	stateMachine.Register("ZPT_Transaction_GetAttributes", this.transactionGetAttributes)

	engine := exec.NewExecutionEngine(
		this.Tx,
		new(util.ECDsaCrypto),
		stateMachine,
	)

	contract := &states.Contract{}
	contract.Deserialize(bytes.NewBuffer(this.Code))
	addr := contract.Address
	dpcode, err := this.GetContractCodeFromAddress(addr)
	if err != nil {
		errStr := err.Error()
		fmt.Printf("err %s %s\n", errStr, addr.ToHexString())
		return nil, fmt.Errorf("get contract  error: %s", addr.ToHexString())
	}
	ccode := dpcode

	var caller common.Address
	if this.ContextRef.CallingContext() == nil {
		caller = common.Address{}
	} else {
		caller = this.ContextRef.CallingContext().ContractAddress
	}
	this.ContextRef.PushContext(&context.Context{ContractAddress: contract.Address})
	res, err := engine.Call(caller, ccode, contract.Method, contract.Args, contract.Version)

	if err != nil {
		return nil, err
	}

	//get the return message
	result, err := engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil {
		return nil, err
	}

	this.ContextRef.PopContext()
	this.ContextRef.PushNotifications(this.Notifications)
	return result, nil
}

func (this *WasmVmService) marshalEmbeddedParams(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[marshalEmbeddedParams]parameter count error while call marshalNativeParams")
	}
	argbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	bytesLen := len(argbytes)
	args := make([]interface{}, bytesLen/8)
	icount := 0
	for i := 0; i < bytesLen; i += 8 {
		tmpBytes := argbytes[i : i+8]
		ptype, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpBytes[:4])))
		if err != nil {
			return false, err
		}
		pvalue, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpBytes[4:8])))
		if err != nil {
			return false, err
		}
		switch strings.ToLower(util.TrimBuffToString(ptype)) {
		case "string":
			args[icount] = util.TrimBuffToString(pvalue)
		case "int":
			args[icount], err = strconv.Atoi(util.TrimBuffToString(pvalue))
			if err != nil {
				return false, err
			}
		case "int64":
			args[icount], err = strconv.ParseInt(util.TrimBuffToString(pvalue), 10, 64)
			if err != nil {
				return false, err
			}
		default:
			args[icount] = util.TrimBuffToString(pvalue)
		}
		icount++
	}
	builder := simulator.NewParamsBuilder(bytes.NewBuffer(nil))
	err = buildEmbeddedParamInter(builder, []interface{}{args})
	if err != nil {
		return false, err
	}
	embedargs := builder.ToArray()
	idx, err := vm.SetPointerMemory(embedargs)
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil

}

// marshalNativeParams
// make parameter bytes for call native contract
func (this *WasmVmService) marshalNativeParams(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[callContract]parameter count error while call marshalNativeParams")
	}

	transferbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	//transferbytes is a nested struct with states.Transfer
	//type Transfers struct {
	//	States  []*State		   -------->i32 pointer 4 bytes
	//}
	if len(transferbytes) != 4 {
		return false, errors.NewErr("[callContract]parameter format error while call marshalNativeParams")
	}
	transfer := &nstates.Transfers{}

	statesAddr := binary.LittleEndian.Uint32(transferbytes[:4])
	statesbytes, err := vm.GetPointerMemory(uint64(statesAddr))
	if err != nil {
		return false, err
	}

	//statesbytes is slice of struct with states.
	//type State struct {
	//	From    common.Address  -------->i32 pointer 4 bytes
	//	To      common.Address  -------->i32 pointer 4 bytes
	//	Value   *big.Int        -------->i64 8 bytes
	//}
	//total is 4 + 4 + 8 = 24 bytes
	statecnt := len(statesbytes) / 16
	states := make([]nstates.State, statecnt)

	for i := 0; i < statecnt; i++ {
		tmpbytes := statesbytes[i*16 : (i+1)*16]
		state := nstates.State{}
		fromAddessBytes, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpbytes[:4])))
		if err != nil {
			return false, err
		}
		fromAddress, err := common.AddressFromBase58(util.TrimBuffToString(fromAddessBytes))
		if err != nil {
			return false, err
		}
		state.From = fromAddress

		toAddressBytes, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpbytes[4:8])))
		if err != nil {
			return false, err
		}
		toAddress, err := common.AddressFromBase58(util.TrimBuffToString(toAddressBytes))
		state.To = toAddress
		//tmpbytes[12:16] is padding
		amount := binary.LittleEndian.Uint64(tmpbytes[8:])
		state.Value = amount
		states[i] = state

	}

	transfer.States = states
	tbytes := new(bytes.Buffer)
	transfer.Serialize(tbytes)

	result, err := vm.SetPointerMemory(tbytes.Bytes())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(result))
	return true, nil
}

// callContract
// need 4 parameters
//0: contract address
//1: contract code
//2: method name
//3: args
/*func (this *WasmVmService) callContract(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 4 {
		return false, errors.NewErr("[callContract]parameter count error while call readMessage")
	}
	var contractAddress common.Address
	var contractBytes []byte
	//get contract address
	contractAddressIdx := params[0]
	addr, err := vm.GetPointerMemory(contractAddressIdx)
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract address failed:" + err.Error())
	}

	if addr != nil {
		addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
		if err != nil {
			return false, errors.NewErr("[callContract]get contract address error:" + err.Error())
		}
		contractAddress, err = common.AddressParseFromBytes(addrbytes)
		if err != nil {
			return false, errors.NewErr("[callContract]get contract address error:" + err.Error())
		}

	}

	//get contract code
	codeIdx := params[1]

	offchainContractCode, err := vm.GetPointerMemory(codeIdx)
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract address failed:" + err.Error())
	}
	if offchainContractCode != nil {
		contractBytes, err = common.HexToBytes(util.TrimBuffToString(offchainContractCode))
		if err != nil {
			return false, err

		}
		//compute the offchain code address
		codestring := util.TrimBuffToString(offchainContractCode)
		contractAddress = GetContractAddress(codestring, vmtypes.WASMVM)
	}
	//get method
	methodName, err := vm.GetPointerMemory(params[2])
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract methodName failed:" + err.Error())
	}
	//get args
	arg, err := vm.GetPointerMemory(params[3])

	if err != nil {
		return false, errors.NewErr("[callContract]get Contract arg failed:" + err.Error())
	}
	this.ContextRef.PushContext(&context.Context{
		Code:            vm.VMCode,
		ContractAddress: vm.ContractAddress})
	result, err := this.ContextRef.AppCall(contractAddress, util.TrimBuffToString(methodName), contractBytes, arg)

	this.ContextRef.PopContext()
	if err != nil {
		return false, errors.NewErr("[callContract]AppCall failed:" + err.Error())
	}
	vm.RestoreCtx()
	var res string
	if envCall.GetReturns() {
		if contractAddress[0] == byte(vmtypes.Embed) {
			result = sccommon.ConvertEmbeddedReturnTypes(result)
			switch result.(type) {
			case int:
				res = strconv.Itoa(result.(int))
			case int64:
				res = strconv.FormatInt(result.(int64), 10)
			case string:
				res = result.(string)
			case []byte:
				tmp := result.([]byte)
				if len(tmp) == 1 {
					if tmp[0] == byte(1) {
						res = "true"
					}
					if tmp[0] == byte(0) {
						res = "false"
					}
				} else {
					res = string(result.([]byte))
				}
			default:
				res = fmt.Sprintf("%s", result)
			}

		}
		if contractAddress[0] == byte(vmtypes.Native) {
			bresult := result.(bool)
			if bresult == true {
				res = "true"
			} else {
				res = "false"
			}

		}
		if contractAddress[0] == byte(vmtypes.WASMVM) {
			res = fmt.Sprintf("%s", result)
		}

		idx, err := vm.SetPointerMemory(res)
		if err != nil {
			return false, errors.NewErr("[callContract]SetPointerMemory failed:" + err.Error())
		}
		vm.PushResult(uint64(idx))
	}

	return true, nil
}*/

func (this *WasmVmService) GetContractCodeFromAddress(address common.Address) ([]byte, error) {

	dcode, err := this.Store.GetContractState(address)
	if err != nil {
		return nil, err
	}

	if dcode == nil {
		return nil, errors.NewErr("[GetContractCodeFromAddress] deployed code is nil")
	}

	return dcode.Code, nil

}

func (this *WasmVmService) getContractFromAddr(addr []byte) ([]byte, error) {
	addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
	if err != nil {
		return nil, errors.NewErr("get contract address error")
	}
	contactaddress, err := common.AddressParseFromBytes(addrbytes)
	if err != nil {
		return nil, errors.NewErr("get contract address error")
	}
	dpcode, err := this.GetContractCodeFromAddress(contactaddress)
	if err != nil {
		return nil, errors.NewErr("get contract  error")
	}
	return dpcode, nil
}

//GetContractAddress return contract address
/*func GetContractAddress(code string, vmType vmtypes.VmType) common.Address {
	data, _ := hex.DecodeString(code)
	vmCode := &vmtypes.VmCode{
		VmType: vmType,
		Code:   data,
	}
	return vmCode.AddressFromVmCode()
}*/

//buildEmbeddedParamInter build embedded invoke param code
func buildEmbeddedParamInter(builder *simulator.ParamsBuilder, smartContractParams []interface{}) error {
	//VM load params in reverse order
	for i := len(smartContractParams) - 1; i >= 0; i-- {
		switch v := smartContractParams[i].(type) {
		case bool:
			builder.EmitPushBool(v)
		case int:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int64:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case common.Fixed64:
			builder.EmitPushInteger(big.NewInt(int64(v.GetData())))
		case uint64:
			val := big.NewInt(0)
			builder.EmitPushInteger(val.SetUint64(uint64(v)))
		case string:
			builder.EmitPushByteArray([]byte(v))
		case *big.Int:
			builder.EmitPushInteger(v)
		case []byte:
			builder.EmitPushByteArray(v)
		case []interface{}:
			err := buildEmbeddedParamInter(builder, v)
			if err != nil {
				return err
			}
			builder.EmitPushInteger(big.NewInt(int64(len(v))))
			builder.Emit(simulator.PACK)
		default:
			return fmt.Errorf("unsupported param:%s", v)
		}
	}
	return nil
}
