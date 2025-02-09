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

package embed

import (
	"github.com/imZhuFei/zeepin/core/payload"
	"github.com/imZhuFei/zeepin/core/types"
	vm "github.com/imZhuFei/zeepin/embed/simulator"
	"github.com/imZhuFei/zeepin/errors"
)

func validatorAttribute(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorAttribute] Too few input parameters ")
	}
	d, err := vm.PeekInteropInterface(engine)
	if err != nil {
		return err
	}
	if d == nil {
		return errors.NewErr("[validatorAttribute] Pop txAttribute nil!")
	}
	_, ok := d.(*types.TxAttribute)
	if ok == false {
		return errors.NewErr("[validatorAttribute] Wrong type!")
	}
	return nil
}

func validatorBlock(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[Block] Too few input parameters ")
	}
	if _, err := peekBlock(engine); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[validatorBlock] Validate block fail!")
	}
	return nil
}

func validatorBlockTransaction(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 2 {
		return errors.NewErr("[validatorBlockTransaction] Too few input parameters ")
	}
	block, err := peekBlock(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[validatorBlockTransaction] Validate block fail!")
	}
	item, err := vm.PeekNBigInt(1, engine)
	if err != nil {
		return err
	}
	index := int(item.Int64())
	if index < 0 {
		return errors.NewErr("[validatorBlockTransaction] Pop index invalid!")
	}
	if index >= len(block.Transactions) {
		return errors.NewErr("[validatorBlockTransaction] index invalid!")
	}
	return nil
}

func validatorBlockChainHeader(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorBlockChainHeader] Too few input parameters ")
	}
	return nil
}

func validatorBlockChainBlock(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorBlockChainBlock] Too few input parameters ")
	}
	return nil
}

func validatorBlockChainTransaction(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorBlockChainTransaction] Too few input parameters ")
	}
	return nil
}

func validatorBlockChainContract(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorBlockChainContract] Too few input parameters ")
	}
	return nil
}

func validatorHeader(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorHeader] Too few input parameters ")
	}
	item, err := vm.PeekInteropInterface(engine)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.NewErr("[validatorHeader] Blockdata is nil!")
	}
	return nil
}

func validatorTransaction(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorTransaction] Too few input parameters ")
	}
	item, err := vm.PeekInteropInterface(engine)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.NewErr("[validatorTransaction] Blockdata is nil!")
	}
	_, ok := item.(*types.Transaction)
	if !ok {
		return errors.NewErr("[validatorTransaction] Transaction wrong type!")
	}
	return nil
}

func validatorGetCode(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorGetCode] Too few input parameters ")
	}
	item, err := vm.PeekInteropInterface(engine)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.NewErr("[validatorGetCode] Contract is nil!")
	}
	deploy, ok := item.(*payload.DeployCode)
	if !ok || deploy == nil {
		return errors.NewErr("[validatorGetCode] DeployCode wrong type!")
	}
	return nil
}

func validatorCheckWitness(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorCheckWitness] Too few input parameters ")
	}
	return nil
}

func validatorNotify(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorNotify] Too few input parameters ")
	}
	return nil
}

func validatorLog(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorLog] Too few input parameters ")
	}
	return nil
}

func validatorSerialize(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorSerialize] Too few input parameters ")
	}
	return nil
}

func validatorDeserialize(engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[validatorDeSerialize] Too few input parameters ")
	}
	return nil
}

func peekBlock(engine *vm.ExecutionEngine) (*types.Block, error) {
	d, err := vm.PeekInteropInterface(engine)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, errors.NewErr("[Block] Pop blockdata nil!")
	}
	block, ok := d.(*types.Block)
	if !ok {
		return nil, errors.NewErr("[Block] Wrong type!")
	}
	return block, nil
}
