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

package test

import (
	"os"
	"testing"

	"github.com/imZhuFei/zeepin/core/types"
	. "github.com/imZhuFei/zeepin/smartcontract"
	"github.com/stretchr/testify/assert"
)

func TestInfiniteLoopCrash(t *testing.T) {
	evilBytecode := []byte(" e\xff\u007f\xffhm\xb7%\xa7AAAAAAAAAAAAAAAC\xef\xed\x04INVERT\x95ve")
	dbFile := "test"
	defer func() {
		os.RemoveAll(dbFile)
	}()
	//testLevelDB, err := leveldbstore.NewLevelDBStore(dbFile)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//store := statestore.NewMemDatabase()
	//testBatch := statestore.NewStateStoreBatch(store, testLevelDB)
	config := &Config{
		Time:   10,
		Height: 10,
		Tx:     &types.Transaction{},
	}
	//cache := storage.NewCloneCache(testBatch)
	sc := SmartContract{
		Config:     config,
		Gas:        10000,
		CloneCache: nil,
	}
	engine, err := sc.NewExecuteEngine(evilBytecode)
	if err != nil {
		t.Fatal(err)
	}
	_, err = engine.Invoke()
	assert.Equal(t, "[EmbeddedService] vm execute error!: the biginteger over max size 32bit", err.Error())
}
