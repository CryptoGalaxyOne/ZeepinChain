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

package consensus

import (
	"github.com/imZhuFei/zeepin/account"
	"github.com/imZhuFei/zeepin/common/log"
	"github.com/imZhuFei/zeepin/consensus/dbft"
	"github.com/imZhuFei/zeepin/consensus/solo"
	"github.com/imZhuFei/zeepin/consensus/vbft"
	"github.com/ontio/ontology-eventbus/actor"
)

type ConsensusService interface {
	Start() error
	Halt() error
	GetPID() *actor.PID
}

const (
	CONSENSUS_DBFT = "dbft"
	CONSENSUS_SOLO = "solo"
	CONSENSUS_VBFT = "gbft"
)

func NewConsensusService(consensusType string, account *account.Account, txpool *actor.PID, ledger *actor.PID, p2p *actor.PID) (ConsensusService, error) {
	if consensusType == "" {
		consensusType = CONSENSUS_DBFT
	}
	var consensus ConsensusService
	var err error
	switch consensusType {
	case CONSENSUS_DBFT:
		consensus, err = dbft.NewDbftService(account, txpool, p2p)
	case CONSENSUS_SOLO:
		consensus, err = solo.NewSoloService(account, txpool)
	case CONSENSUS_VBFT:
		consensus, err = vbft.NewVbftServer(account, txpool, p2p)
	}
	log.Infof("ConsensusType:%s", consensusType)
	return consensus, err
}
