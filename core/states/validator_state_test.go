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
package states

import (
	"testing"

	"bytes"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
)

func TestValidatorState_Deserialize_Serialize(t *testing.T) {
	_, pubKey, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)

	vs := ValidatorState{
		StateBase: StateBase{(byte)(1)},
		PublicKey: pubKey,
	}

	buf := bytes.NewBuffer(nil)
	vs.Serialize(buf)
	bs := buf.Bytes()

	var vs2 ValidatorState
	vs2.Deserialize(buf)
	assert.Equal(t, vs, vs2)

	buf = bytes.NewBuffer(bs[:len(bs)-1])
	err := vs2.Deserialize(buf)
	assert.NotNil(t, err)
}
