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

package cmd

import (
	"fmt"

	"github.com/imZhuFei/zeepin/cmd/utils"
	"github.com/imZhuFei/zeepin/common"
	"github.com/imZhuFei/zeepin/common/config"
	"github.com/imZhuFei/zeepin/common/log"
	"github.com/imZhuFei/zeepin/smartcontract/service/native/governance"
	"github.com/urfave/cli"
)

func SetZeepinChainConfig(ctx *cli.Context) (*config.ZeepinChainConfig, error) {
	cfg := config.DefConfig
	err := setGenesis(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("setGenesis error:%s", err)
	}
	setCommonConfig(ctx, cfg.Common)
	setConsensusConfig(ctx, cfg.Consensus)
	setP2PNodeConfig(ctx, cfg.P2PNode)
	setRpcConfig(ctx, cfg.Rpc)
	setRestfulConfig(ctx, cfg.Restful)
	setWebSocketConfig(ctx, cfg.Ws)
	if cfg.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		cfg.Ws.EnableHttpWs = true
		cfg.Restful.EnableHttpRestful = true
		cfg.Consensus.EnableConsensus = true
		cfg.P2PNode.NetworkId = config.NETWORK_ID_SOLO_NET
		cfg.P2PNode.NetworkName = config.GetNetworkName(cfg.P2PNode.NetworkId)
		cfg.P2PNode.NetworkMagic = config.GetNetworkMagic(cfg.P2PNode.NetworkId)
		cfg.Common.GasPrice = 0
	}
	if cfg.P2PNode.NetworkId == config.NETWORK_ID_MAIN_NET ||
		cfg.P2PNode.NetworkId == config.NETWORK_ID_POLARIS_NET {
		defNetworkId, err := cfg.GetDefaultNetworkId()
		if err != nil {
			return nil, fmt.Errorf("GetDefaultNetworkId error:%s", err)
		}
		if defNetworkId != cfg.P2PNode.NetworkId {
			cfg.P2PNode.NetworkId = defNetworkId
			cfg.P2PNode.NetworkMagic = config.GetNetworkMagic(defNetworkId)
			cfg.P2PNode.NetworkName = config.GetNetworkName(defNetworkId)
		}
	}
	return cfg, nil
}

func setGenesis(ctx *cli.Context, cfg *config.ZeepinChainConfig) error {
	netWorkId := ctx.GlobalInt(utils.GetFlagName(utils.NetworkIdFlag))
	switch netWorkId {
	case config.NETWORK_ID_MAIN_NET:
		cfg.Genesis = config.MainNetConfig
	case config.NETWORK_ID_POLARIS_NET:
		cfg.Genesis = config.PolarisConfig
	}

	if ctx.GlobalBool(utils.GetFlagName(utils.EnableTestModeFlag)) {
		cfg.Genesis.ConsensusType = config.CONSENSUS_TYPE_SOLO
		cfg.Genesis.SOLO.GenBlockTime = ctx.Uint(utils.GetFlagName(utils.TestModeGenBlockTimeFlag))
		if cfg.Genesis.SOLO.GenBlockTime <= 1 {
			cfg.Genesis.SOLO.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
		return nil
	}

	if !ctx.IsSet(utils.GetFlagName(utils.ConfigFlag)) {
		return nil
	}

	genesisFile := ctx.GlobalString(utils.GetFlagName(utils.ConfigFlag))
	if !common.FileExisted(genesisFile) {
		return nil
	}

	newGenesisCfg := config.NewGenesisConfig()
	err := utils.GetJsonObjectFromFile(genesisFile, newGenesisCfg)
	if err != nil {
		return err
	}
	cfg.Genesis = newGenesisCfg
	log.Infof("Load genesis config:%s", genesisFile)

	switch cfg.Genesis.ConsensusType {
	case config.CONSENSUS_TYPE_DBFT:
		if len(cfg.Genesis.DBFT.Bookkeepers) < config.DBFT_MIN_NODE_NUM {
			return fmt.Errorf("DBFT consensus at least need %d bookkeepers in config", config.DBFT_MIN_NODE_NUM)
		}
		if cfg.Genesis.DBFT.GenBlockTime <= 0 {
			cfg.Genesis.DBFT.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
	case config.CONSENSUS_TYPE_VBFT:
		err = governance.CheckVBFTConfig(cfg.Genesis.GBFT)
		if err != nil {
			return fmt.Errorf("GBFT config error %v", err)
		}
		if len(cfg.Genesis.GBFT.Peers) < config.VBFT_MIN_NODE_NUM {
			return fmt.Errorf("GBFT consensus at least need %d peers in config", config.VBFT_MIN_NODE_NUM)
		}
	default:
		return fmt.Errorf("Unknow consensus:%s", cfg.Genesis.ConsensusType)
	}

	return nil
}

func setCommonConfig(ctx *cli.Context, cfg *config.CommonConfig) {
	cfg.LogLevel = ctx.GlobalUint(utils.GetFlagName(utils.LogLevelFlag))
	cfg.EnableEventLog = !ctx.GlobalBool(utils.GetFlagName(utils.DisableEventLogFlag))
	cfg.GasLimit = ctx.GlobalUint64(utils.GetFlagName(utils.GasLimitFlag))
	cfg.GasPrice = ctx.GlobalUint64(utils.GetFlagName(utils.GasPriceFlag))
	cfg.DataDir = ctx.GlobalString(utils.GetFlagName(utils.DataDirFlag))
}

func setConsensusConfig(ctx *cli.Context, cfg *config.ConsensusConfig) {
	cfg.EnableConsensus = ctx.GlobalBool(utils.GetFlagName(utils.EnableConsensusFlag))
	cfg.MaxTxInBlock = ctx.GlobalUint(utils.GetFlagName(utils.MaxTxInBlockFlag))
}

func setP2PNodeConfig(ctx *cli.Context, cfg *config.P2PNodeConfig) {
	cfg.NetworkId = uint32(ctx.GlobalUint(utils.GetFlagName(utils.NetworkIdFlag)))
	cfg.NetworkMagic = config.GetNetworkMagic(cfg.NetworkId)
	cfg.NetworkName = config.GetNetworkName(cfg.NetworkId)
	cfg.NodePort = ctx.GlobalUint(utils.GetFlagName(utils.NodePortFlag))
	cfg.NodeConsensusPort = ctx.GlobalUint(utils.GetFlagName(utils.ConsensusPortFlag))
	cfg.DualPortSupport = ctx.GlobalBool(utils.GetFlagName(utils.DualPortSupportFlag))
	cfg.ReservedPeersOnly = ctx.GlobalBool(utils.GetFlagName(utils.ReservedPeersOnlyFlag))
	cfg.MaxConnInBound = ctx.GlobalUint(utils.GetFlagName(utils.MaxConnInBoundFlag))
	cfg.MaxConnOutBound = ctx.GlobalUint(utils.GetFlagName(utils.MaxConnOutBoundFlag))
	cfg.MaxConnInBoundForSingleIP = ctx.GlobalUint(utils.GetFlagName(utils.MaxConnInBoundForSingleIPFlag))

	rsvfile := ctx.GlobalString(utils.GetFlagName(utils.ReservedPeersFileFlag))
	if cfg.ReservedPeersOnly {
		if !common.FileExisted(rsvfile) {
			log.Infof("file %s not exist\n", rsvfile)
			return
		}
		err := utils.GetJsonObjectFromFile(rsvfile, &cfg.ReservedCfg)
		if err != nil {
			log.Errorf("Get ReservedCfg error:%s", err)
			return
		}
		for i := 0; i < len(cfg.ReservedCfg.ReservedPeers); i++ {
			log.Info("reserved addr: " + cfg.ReservedCfg.ReservedPeers[i])
		}
		for i := 0; i < len(cfg.ReservedCfg.MaskPeers); i++ {
			log.Info("mask addr: " + cfg.ReservedCfg.MaskPeers[i])
		}
	}

}

func setRpcConfig(ctx *cli.Context, cfg *config.RpcConfig) {
	cfg.EnableHttpJsonRpc = !ctx.Bool(utils.GetFlagName(utils.RPCDisabledFlag))
	cfg.HttpJsonPort = ctx.GlobalUint(utils.GetFlagName(utils.RPCPortFlag))
	cfg.HttpLocalPort = ctx.GlobalUint(utils.GetFlagName(utils.RPCLocalProtFlag))
}

func setRestfulConfig(ctx *cli.Context, cfg *config.RestfulConfig) {
	cfg.EnableHttpRestful = ctx.GlobalBool(utils.GetFlagName(utils.RestfulEnableFlag))
	cfg.HttpRestPort = ctx.GlobalUint(utils.GetFlagName(utils.RestfulPortFlag))
}

func setWebSocketConfig(ctx *cli.Context, cfg *config.WebSocketConfig) {
	cfg.EnableHttpWs = ctx.GlobalBool(utils.GetFlagName(utils.WsEnabledFlag))
	cfg.HttpWsPort = ctx.GlobalUint(utils.GetFlagName(utils.WsPortFlag))
}

func SetRpcPort(ctx *cli.Context) {
	if ctx.IsSet(utils.GetFlagName(utils.RPCPortFlag)) {
		config.DefConfig.Rpc.HttpJsonPort = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	}
}
