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

package utils

import (
	"bytes"

	"github.com/imZhuFei/zeepin/common/serialization"
	cstates "github.com/imZhuFei/zeepin/core/states"
	scommon "github.com/imZhuFei/zeepin/core/store/common"
	"github.com/imZhuFei/zeepin/errors"
	"github.com/imZhuFei/zeepin/smartcontract/service/native"
)

func GetStorageItem(native *native.NativeService, key []byte) (*cstates.StorageItem, error) {
	store, err := native.CloneCache.Get(scommon.ST_STORAGE, key)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageItem] storage error!")
	}
	if store == nil {
		return nil, nil
	}
	item, ok := store.(*cstates.StorageItem)
	if !ok {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageItem] instance doesn't StorageItem!")
	}
	return item, nil
}

func GetStorageUInt64(native *native.NativeService, key []byte) (uint64, error) {
	item, err := GetStorageItem(native, key)
	if err != nil {
		return 0, err
	}
	if item == nil {
		return 0, nil
	}
	v, err := serialization.ReadUint64(bytes.NewBuffer(item.Value))
	if err != nil {
		return 0, err
	}
	return v, nil
}

func GetStorageUInt32(native *native.NativeService, key []byte) (uint32, error) {
	item, err := GetStorageItem(native, key)
	if err != nil {
		return 0, err
	}
	if item == nil {
		return 0, nil
	}
	v, err := serialization.ReadUint32(bytes.NewBuffer(item.Value))
	if err != nil {
		return 0, err
	}
	return v, nil
}

func GenUInt64StorageItem(value uint64) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	serialization.WriteUint64(bf, value)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func GenUInt32StorageItem(value uint32) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	serialization.WriteUint32(bf, value)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func PutBytes(native *native.NativeService, key []byte, value []byte) {
	native.CloneCache.Add(scommon.ST_STORAGE, key, &cstates.StorageItem{Value: value})
}
