package script

import (
	"encoding/binary"
)

// nft
func decodeNFT(scriptLen int, scriptPk []byte, txo *TxoData) bool {
	dataLen := 0
	protoVersionLen := 0
	genesisIdLen := 76 // nft v2
	sensibleIdLen := 36
	useTokenIdHash := false
	if scriptPk[scriptLen-89-76-1-1-1] == OP_RETURN &&
		scriptPk[scriptLen-89-76-1-1] == 0x4c &&
		scriptPk[scriptLen-89-76-1] == 165 {
		// nft v3
		// <nft data> = <metaid_outpoint(36 bytes)> + <is_genesis(1 byte)> + <address(20 bytes)> + <totalSupply(8 bytes) + <tokenIndex(8 bytes)> + <genesisHash<20 bytes>) + <RABIN_PUBKEY_HASH_ARRAY_HASH(20 bytes)> + <sensibleID(36 bytes)> + <protoVersion(4 bytes)> + <protoType(4 bytes)> + <protoFlag(8 bytes)>
		dataLen = 1 + 1 + 36 + 1 + 20 + 8 + 8 + 20 + 20 + 36 + 4 + 4 + 8 // 0x4c + pushdata + data
		protoVersionLen = 4
		useTokenIdHash = true
	} else if scriptPk[scriptLen-85-76-1-1-1] == OP_RETURN &&
		scriptPk[scriptLen-85-76-1-1] == 0x4c &&
		scriptPk[scriptLen-85-76-1] == 161 {
		// nft v2
		// <nft data> = <metaid_outpoint(36 bytes)> + <is_genesis(1 byte)> + <address(20 bytes)> + <totalSupply(8 bytes) + <tokenIndex(8 bytes)> + <genesisHash<20 bytes>) + <RABIN_PUBKEY_HASH_ARRAY_HASH(20 bytes)> + <sensibleID(36 bytes)> + <protoType(4 bytes)> + <protoFlag(8 bytes)>
		dataLen = 1 + 1 + 1 + 36 + 1 + 20 + 8 + 8 + 20 + 20 + 36 + 4 + 8 // opreturn + 0x4c + pushdata + data
		protoVersionLen = 0
		useTokenIdHash = false
	} else {
		return false
	}

	protoTypeOffset := scriptLen - 8 - 4
	sensibleOffset := protoTypeOffset - protoVersionLen - sensibleIdLen

	genesisOffset := protoTypeOffset - protoVersionLen - genesisIdLen
	tokenIndexOffset := genesisOffset - 8
	tokenSupplyOffset := tokenIndexOffset - 8
	addressOffset := tokenSupplyOffset - 20
	isGenesisOffset := addressOffset - 1
	metaOutputIndexOffset := isGenesisOffset - 4
	metaTxIdOffset := metaOutputIndexOffset - 32

	txo.CodeType = CodeType_NFT
	txo.CodeHash = GetHash160(scriptPk[:scriptLen-dataLen])

	txo.SensibleId = make([]byte, sensibleIdLen)
	copy(txo.SensibleId, scriptPk[sensibleOffset:sensibleOffset+sensibleIdLen])

	if useTokenIdHash {
		// GenesisId is tokenIdHash
		txo.GenesisId = GetHash160(scriptPk[genesisOffset : genesisOffset+genesisIdLen])
	} else {
		// for search: codehash + genesis
		txo.GenesisId = txo.SensibleId
	}

	txo.MetaOutputIndex = binary.LittleEndian.Uint32(scriptPk[metaOutputIndexOffset : metaOutputIndexOffset+4])
	txo.MetaTxId = ReverseBytes(scriptPk[metaTxIdOffset : metaTxIdOffset+32])

	txo.TokenSupply = binary.LittleEndian.Uint64(scriptPk[tokenSupplyOffset : tokenSupplyOffset+8])
	txo.TokenIndex = binary.LittleEndian.Uint64(scriptPk[tokenIndexOffset : tokenIndexOffset+8])
	if scriptPk[isGenesisOffset] == 1 {
		txo.TokenIndex = txo.TokenSupply
	}

	txo.AddressPkh = make([]byte, 20)
	copy(txo.AddressPkh, scriptPk[addressOffset:addressOffset+20])
	return true
}