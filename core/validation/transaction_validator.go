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

package validation

import (
	"errors"
	"fmt"

	"github.com/imZhuFei/zeepin/common"
	"github.com/imZhuFei/zeepin/common/constants"
	"github.com/imZhuFei/zeepin/common/log"
	"github.com/imZhuFei/zeepin/core/ledger"
	"github.com/imZhuFei/zeepin/core/payload"
	"github.com/imZhuFei/zeepin/core/signature"
	"github.com/imZhuFei/zeepin/core/types"
	ontErrors "github.com/imZhuFei/zeepin/errors"
)

// VerifyTransaction verifys received single transaction
func VerifyTransaction(tx *types.Transaction) ontErrors.ErrCode {
	if err := checkTransactionSignatures(tx); err != nil {
		log.Info("transaction verify error:", err)
		return ontErrors.ErrVerifySignature
	}

	if err := checkTransactionPayload(tx); err != nil {
		log.Warn("[VerifyTransaction],", err)
		return ontErrors.ErrTransactionPayload
	}

	return ontErrors.ErrNoError
}

func VerifyTransactionWithLedger(tx *types.Transaction, ledger *ledger.Ledger) ontErrors.ErrCode {
	//TODO: replay check
	return ontErrors.ErrNoError
}

func checkTransactionSignatures(tx *types.Transaction) error {
	hash := tx.Hash()

	lensig := len(tx.Sigs)
	if lensig > constants.TX_MAX_SIG_SIZE {
		return fmt.Errorf("transaction signature number %d execced %d", lensig, constants.TX_MAX_SIG_SIZE)
	}

	address := make(map[common.Address]bool, len(tx.Sigs))
	for _, sig := range tx.Sigs {
		m := int(sig.M)
		kn := len(sig.PubKeys)
		sn := len(sig.SigData)

		if kn > constants.MULTI_SIG_MAX_PUBKEY_SIZE || sn < m || m > kn || m <= 0 {
			return errors.New("wrong tx sig param length")
		}

		if kn == 1 {
			err := signature.Verify(sig.PubKeys[0], hash[:], sig.SigData[0])
			if err != nil {
				return errors.New("signature verification failed")
			}

			address[types.AddressFromPubKey(sig.PubKeys[0])] = true
		} else {
			if err := signature.VerifyMultiSignature(hash[:], sig.PubKeys, m, sig.SigData); err != nil {
				return err
			}

			addr, err := types.AddressFromMultiPubKeys(sig.PubKeys, m)
			if err != nil {
				return err
			}
			address[addr] = true
		}
	}

	// check payer in address
	if address[tx.Payer] == false {
		return errors.New("signature missing for payer: " + tx.Payer.ToBase58())
	}
	addrList := make([]common.Address, 0, len(address))
	for addr := range address {
		addrList = append(addrList, addr)
	}
	tx.SignedAddr = addrList

	return nil
}

func checkTransactionPayload(tx *types.Transaction) error {

	switch pld := tx.Payload.(type) {
	case *payload.DeployCode:
		return nil
	case *payload.InvokeCode:
		return nil
	default:
		return errors.New(fmt.Sprint("[txValidator], unimplemented transaction payload type.", pld))
	}
	return nil
}
