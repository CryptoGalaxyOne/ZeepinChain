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

import "testing"

func TestNewWasmStateMachine(t *testing.T) {
	sm := NewWasmStateMachine()
	if sm == nil {
		t.Fatal("NewWasmStateMachine should return a non nil state machine")
	}

	if sm.WasmStateReader == nil {
		t.Fatal("NewWasmStateMachine should return a non nil state reader")
	}

	if !sm.Exists("ContractLogDebug") {
		t.Error("NewWasmStateMachine should has ContractLogDebug service")
	}

	if !sm.Exists("ContractLogInfo") {
		t.Error("NewWasmStateMachine should has ContractLogInfo service")
	}

	if !sm.Exists("ContractLogError") {
		t.Error("NewWasmStateMachine should has ContractLogError service")
	}
}
