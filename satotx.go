package script

import (
	"bytes"
	"encoding/binary"
)

func hasMetaContractFlag(pkScript []byte) bool {
	if len(pkScript) < 6 {
		return false
	}
	script := pkScript[:len(pkScript)-5]
	return bytes.HasSuffix(script, []byte("metacontract"))
}

func DecodeMvcTxo(pkScript []byte, txo *TxoData) bool {
	scriptLen := len(pkScript)
	if scriptLen < 1024 {
		return false
	}

	ret := false
	if hasMetaContractFlag(pkScript) {
		// 偏移长度5，偏移metacontract 12
		protoTypeOffset := scriptLen - 5 - 12 - 4
		protoType := binary.LittleEndian.Uint32(pkScript[protoTypeOffset : protoTypeOffset+4])

		switch protoType {
		case CodeType_FT:
			ret = decodeMvcFT(scriptLen, pkScript, txo)

		case CodeType_UNIQUE:
			ret = decodeMvcUnique(scriptLen, pkScript, txo)

		case CodeType_NFT:
			ret = decodeMvcNFT(scriptLen, pkScript, txo)

		case CodeType_NFT_SELL:
			ret = decodeNFTSell(scriptLen, pkScript, txo)
		case CodeType_NFT_SELL_V2:
			ret = decodeNFTSellV2(scriptLen, pkScript, txo)
		case CodeType_NFT_AUCTION:
			ret = decodeNFTAuction(scriptLen, pkScript, txo)
		default:
			ret = false
		}
		return ret
	}

	if pkScript[scriptLen-1] < 2 &&
		pkScript[scriptLen-37-1] == 37 &&
		pkScript[scriptLen-37-1-40-1] == 40 &&
		pkScript[scriptLen-37-1-40-1-1] == OP_RETURN {
		ret = decodeNFTIssue(scriptLen, pkScript, txo)

	} else if pkScript[scriptLen-1] == 1 &&
		pkScript[scriptLen-61-1] == 61 &&
		pkScript[scriptLen-61-1-40-1] == 40 &&
		pkScript[scriptLen-61-1-40-1-1] == OP_RETURN {
		ret = decodeNFTTransfer(scriptLen, pkScript, txo)
	}

	return ret
}

func ExtractPkScriptForTxo(pkScript, scriptType []byte) (txo *TxoData) {
	txo = &TxoData{}

	if len(pkScript) == 0 {
		return txo
	}

	if isPubkeyHash(scriptType) {
		txo.HasAddress = true
		copy(txo.AddressPkh[:], pkScript[3:23])
		return txo
	}

	if isPayToScriptHash(scriptType) {
		txo.HasAddress = true
		copy(txo.AddressPkh[:], GetHash160(pkScript[2:len(pkScript)-1]))
		return txo
	}

	if isPubkey(scriptType) {
		txo.HasAddress = true
		copy(txo.AddressPkh[:], GetHash160(pkScript[1:len(pkScript)-1]))
		return txo
	}

	// if isMultiSig(scriptType) {
	// 	return pkScript[:]
	// }

	if IsOpreturn(scriptType) {
		if hasMetaContractFlag(pkScript) {
			txo.CodeType = CodeType_SENSIBLE
		}
		return txo
	}

	DecodeMvcTxo(pkScript, txo)

	return txo
}

func GetLockingScriptType(pkScript []byte) (scriptType []byte) {
	length := len(pkScript)
	if length == 0 {
		return
	}
	scriptType = make([]byte, 0)

	lenType := 0
	p := uint(0)
	e := uint(length)

	for p < e && lenType < 32 {
		c := pkScript[p]
		if 0 < c && c < 0x4f {
			cnt, cntsize := SafeDecodeVarIntForScript(pkScript[p:])
			p += cnt + cntsize
		} else {
			p += 1
		}
		scriptType = append(scriptType, c)
		lenType += 1
	}
	return
}
