const (
	VERBOSE = 0
	RAWDATA = 0
)

init = {
	global filterFlashVer = BPL_FILTER["flashVer"]
	global filterReqMode = BPL_FILTER["reqMode"]
	global filterPlay = (filterReqMode == "play")
	global filterDir = BPL_FILTER["dir"]

	if filterDir != undefined {
		if filterDir != BPL_DIRECTION {
			do exit(0)
		}
	}

	global fmsKey = bytes.from([
		0x47, 0x65, 0x6e, 0x75, 0x69, 0x6e, 0x65, 0x20,
		0x41, 0x64, 0x6f, 0x62, 0x65, 0x20, 0x46, 0x6c,
		0x61, 0x73, 0x68, 0x20, 0x4d, 0x65, 0x64, 0x69,
		0x61, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
		0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Media Server 001
		0xf0, 0xee, 0xc2, 0x4a, 0x80, 0x68, 0xbe, 0xe8,
		0x2e, 0x00, 0xd0, 0xd1, 0x02, 0x9e, 0x7e, 0x57,
		0x6e, 0xec, 0x5d, 0x2d, 0x29, 0x80, 0x6f, 0xab,
		0x93, 0xb8, 0xe6, 0x36, 0xcf, 0xeb, 0x31, 0xae,
	])
	global fpKey = bytes.from([
		0x47, 0x65, 0x6E, 0x75, 0x69, 0x6E, 0x65, 0x20,
		0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x46, 0x6C,
		0x61, 0x73, 0x68, 0x20, 0x50, 0x6C, 0x61, 0x79,
		0x65, 0x72, 0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Player 001
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8,
		0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
		0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	])

	if BPL_DIRECTION == "REQ" {
		global handshakeKey = fpKey[:30]
	} else {
		global handshakeKey = fmsKey[:36]
	}

	global lastMsgs = mkmap("int:var")
	global chunksize = 128
	global objectend = errors.new("object end")
	global limitTypes = {
		0: "Hard",
		1: "Soft",
		2: "Dynamic",
	}

	global audioFormats = {
		0:  "Linear PCM",
		1:  "ADPCM",
		2:  "MP3",
		3:  "Linear PCM le",
		4:  "Nellymoser 16kHz",
		5:  "Nellymoser 8kHz",
		6:  "Nellymoser",
		7:  "G711 A-law",
		8:  "G711 mu-law",
		9:  "Reserved",
		10: "AAC",
		11: "Speex",
		12: "MP3 8kHz",
		13: "Device Specific",
	}
	global audioRates = ["5.5 kHz", "11 kHz", "22 kHz", "44 kHz"]
	global audioBits = ["8 bits", "16 bits"]
	global audioChannels = ["Mono", "Stereo"]

	global videoTypes = {
		1: "AVC keyframe",
		2: "AVC inter frame",
		3: "H.263 disposable inter frame",
		4: "generated keyframe",
		5: "video info/command frame",
	}
	global videoCodecs = {
		1: "JPEG",
		2: "H.263",
		3: "screen video",
		4: "On2 VP6",
		5: "On2 VP6 with alpha channel",
		6: "screen video v2",
		7: "AVC",
	}
	global avcTypes = {
		0: "sequence header",
		1: "NALU",
		2: "end of sequence",
	}
}

// --------------------------------------------------------------

AMF0_NUMBER = {
	val float64be
	if VERBOSE == 0 {
		return val
	}
}

AMF0_BOOLEAN = {
	val byte
	if VERBOSE == 0 {
		return byte != 0
	}
}

AMF0_STRING = {
	len uint16be
	val [len]char
	if VERBOSE == 0 {
		return val
	}
}

AMF0_OBJECT_ITEMS = {
	_key AMF0_STRING
	_val AMF0_TYPE
	let items = mkslice("var", 2)
	do set(items, 0, _key, 1, _val)
	if _val != objectend {
		_next AMF0_OBJECT_ITEMS
		let items = append(items, _next.items...)
	}
}

AMF0_OBJECT_NORMAL = {
	val AMF0_OBJECT_ITEMS
	let n = len(val.items)
	return mapFrom(val.items[:n-2]...) // 去掉了最后的 objectend
}

AMF0_OBJECT_VERBOSE = {
	_key AMF0_STRING
	_val AMF0_TYPE
	let items = [{"key": _key, "val": _val}]
	if _val.marker != 0x09 {
		_next AMF0_OBJECT_VERBOSE
		let items = append(items, _next.items...)
	}
}

AMF0_OBJECT = if VERBOSE do AMF0_OBJECT_VERBOSE else AMF0_OBJECT_NORMAL

AMF0_STRICT_ARRAY = {
	len  uint32be
	objs [len]AMF0_TYPE
	if VERBOSE == 0 {
		return objs
	}
}

AMF0_MOVIECLIP = {
	body *byte
	fatal "todo - AMF0_MOVIECLIP"
}

AMF0_NULL = {
	if VERBOSE {
		let val = nil
	} else {
		return nil
	}
}

AMF0_UNDEFINED = {
	if VERBOSE {
		let val = undefined
	} else {
		return undefined
	}
}

AMF0_REFERENCE = {
	reference uint16be
}

AMF0_ECMA_ARRAY = {
	len uint32be
	val AMF0_OBJECT
	if VERBOSE == 0 {
		return val
	}
}

AMF0_OBJECT_END = if VERBOSE do nil else {
	return objectend
}

AMF0_DATE = {
	timestamp float64be
	tz        uint16be
}

AMF0_LONG_STRING = {
	len uint32be
	val [len]char
	if VERBOSE == 0 {
		return val
	}
}

AMF0_UNSUPPORTED = {
	body *byte
	fatal "todo - AMF0_UNSUPPORTED"
}

AMF0_RECORDSET = {
	body *byte
	fatal "todo - AMF0_RECORDSET"
}

AMF0_XML_DOCUMENT = AMF0_LONG_STRING

AMF0_TYPED_OBJECT = {
	type AMF0_STRING
	val  AMF0_OBJECT
}

AMF0_ACMPLUS_OBJECT = { // Switch to AMF3
	body *byte
	fatal "todo - AMF0_ACMPLUS_OBJECT"
}

AMF0_TYPE = {
	marker byte
	case marker {
		0x00: AMF0_NUMBER
		0x01: AMF0_BOOLEAN
		0x02: AMF0_STRING
		0x03: AMF0_OBJECT
		0x04: AMF0_MOVIECLIP
		0x05: AMF0_NULL
		0x06: AMF0_UNDEFINED
		0x07: AMF0_REFERENCE
		0x08: AMF0_ECMA_ARRAY
		0x09: AMF0_OBJECT_END
		0x0a: AMF0_STRICT_ARRAY
		0x0b: AMF0_DATE
		0x0c: AMF0_LONG_STRING
		0x0d: AMF0_UNSUPPORTED
		0x0e: AMF0_RECORDSET
		0x0f: AMF0_XML_DOCUMENT
		0x10: AMF0_TYPED_OBJECT
		0x11: AMF0_ACMPLUS_OBJECT
	}
}

AMF0_CMDDATA = {
	cmd           AMF0_TYPE
	transactionId AMF0_TYPE
	value         *AMF0_TYPE
	if filterFlashVer != undefined {
		if cmd == "connect" {
			if value[0].flashVer != filterFlashVer {
				do exit(0)
			}
		}
	}
	if filterPlay {
		if cmd == "FCPublish" {
			do exit(0)
		}
	}
}

AMF0 = {
	msg *AMF0_TYPE
}

AMF0_CMD = {
	msg AMF0_CMDDATA
}

// --------------------------------------------------------------

AMF3_UNDEFINED = AMF0_UNDEFINED

AMF3_NULL = AMF0_NULL

AMF3_FALSE = {
	if VERBOSE {
		let val = false
	} else {
		return false
	}
}

AMF3_TRUE = {
	if VERBOSE {
		let val = true
	} else {
		return true
	}
}

AMF3_INT = {
	b1 byte
	if b1 & 0x80 {
		let b1 = b1 & 0x7f
		b2 byte
		if b2 & 0x80 {
			let b2 = b2 & 0x7f
			b3 byte
			if b3 & 0x80 {
				let b3 = b3 & 0x7f
				b4 byte
				return (b1 << 22) | (b2 << 15) | (b3 << 8) | b4
			} else {
				return (b1 << 14) | (b2 << 7) | b3
			}
		} else {
			return (b1 << 7) | b2
		}
	} else {
		return int(b1)
	}
}

AMF3_INTEGER_VERBOSE = {
	val AMF3_INT
}

AMF3_INTEGER = if VERBOSE do AMF3_INTEGER_VERBOSE else AMF3_INT

AMF3_DOUBLE = {
	val float64be
	if VERBOSE == 0 {
		return val
	}
}

AMF3_STRING = {
	tag AMF3_INT
	assert (tag & 1) != 0 // reference unsupported
	if tag & 1 {
		val [tag >> 1]char
	}
	if VERBOSE == 0 {
		return val
	}
}

AMF3_XMLDOC = {
	body *byte
}

AMF3_DATE = {
	tag AMF3_INT
	assert (tag & 1) != 0 // reference unsupported
	timestamp float64be
	let tz = tag >> 1
	if VERBOSE == 0 {
		do unset("tag")
	}
}

AMF3_ARRAY = {
	tag AMF3_INT
	assert (tag & 1) != 0 // reference unsupported
	let len = tag >> 1
	body *byte
	fatal "todo - AMF3_ARRAY"
}

AMF3_OBJECT = {
	body *byte
	fatal "todo - AMF3_OBJECT"
}

AMF3_XML = {
	body *byte
	fatal "todo - AMF3_XML"
}

AMF3_BYTE_ARRAY = {
	body *byte
	do println(hex.dump(body))
	fatal "todo - AMF3_BYTE_ARRAY"
}

AMF3_VECTOR_INT = {
	body *byte
	fatal "todo - AMF3_VECTOR_INT"
}

AMF3_VECTOR_UINT = {
	body *byte
	fatal "todo - AMF3_VECTOR_UINT"
}

AMF3_VECTOR_DOUBLE = {
	body *byte
	fatal "todo - AMF3_VECTOR_DOUBLE"
}

AMF3_VECTOR_OBJECT = {
	body *byte
	fatal "todo - AMF3_VECTOR_OBJECT"
}

AMF3_DICTIONARY = {
	body *byte
	fatal "todo - AMF3_DICTIONARY"
}

AMF3_TYPE = {
	marker byte
	case marker {
		0x00: AMF3_UNDEFINED
		0x01: AMF3_NULL
		0x02: AMF3_FALSE
		0x03: AMF3_TRUE
		0x04: AMF3_INTEGER
		0x05: AMF3_DOUBLE
		0x06: AMF3_STRING
		0x07: AMF3_XMLDOC
		0x08: AMF3_DATE
		0x09: AMF3_ARRAY
		0x0a: AMF3_OBJECT
		0x0b: AMF3_XML
		0x0c: AMF3_BYTE_ARRAY
		0x0d: AMF3_VECTOR_INT
		0x0e: AMF3_VECTOR_UINT
		0x0f: AMF3_VECTOR_DOUBLE
		0x10: AMF3_VECTOR_OBJECT
		0x11: AMF3_DICTIONARY
	}
}

AMF3_CMDDATA = {
	cmd           AMF3_TYPE
	transactionId AMF3_TYPE
	value         *AMF3_TYPE
}

AMF3 = {
	msg *AMF3_TYPE
}

AMF3_CMD = {
	msg AMF3_CMDDATA
}

// --------------------------------------------------------------

AudioData = {
	tag byte
	let format = tag >> 4
	let formatKind = audioFormats[format]
	let rate = (tag>>2) & 3
	let rateKind = audioRates[rate]
	let bits = (tag>>1) & 1
	let bitsKind = audioBits[bits]
	let channel = tag & 1
	let channelKind = audioChannels[channel]

	if formatKind == "AAC" {
		aacPacketType byte
		if aacPacketType == 0 {
			let aacPacketTypeKind = "sequence header"
		} else {
			let aacPacketTypeKind = "raw data"
		}
	}
	if RAWDATA != 0 {
		raw *byte
	}
}

Audio = {
	audio AudioData
}

// --------------------------------------------------------------

VideoData = {
	tag byte
	let type = tag >> 4
	let typeKind = videoTypes[type]
	let codec = tag & 0xf
	let codecKind = videoCodecs[codec]

	if codecKind == "AVC" {
		avctype		 byte
		compositionTime uint24be
		let avctypeKind = avcTypes[int(avctype)]
		if avctype == 0 { // sequence header
			configurationVersion byte
			avcProfileIndication byte
			profileCompatibility byte
			avcLevelIndication   byte
			lengthSizeMinusOne   byte
			numOfSPS             byte // SPS = SequenceParameterSets
			bytesOfSPS           uint16be
			dataOfSPS            [bytesOfSPS]byte // SPS包含视频长、宽的信息
			numOfPPS             byte // PPS = PictureParameterSets
			bytesOfPPS           uint16be
			dataOfPPS            [bytesOfPPS]byte
			let lengthSizeMinusOne = (lengthSizeMinusOne & 3) + 1
			let numOfSPS = numOfSPS & 0x1F
			let numOfPPS = numOfPPS & 0x1F
			unknown *byte
		}
	}
	if RAWDATA != 0 {
		raw *byte
	}
}

Video = {
	video VideoData
}

// --------------------------------------------------------------

SetChunkSize = {
	size uint32be
	let chunksize = size
}

Abort = {
	csid uint32be
	let _last = lastMsgs[csid]
	do set(_last, "remain", 0)
}

UserControl = {
	evType uint16be
	evData *byte
}

AckWinsize = {
	winsize uint32be
}

SetPeerBandwidth = {
	winsize   uint32be
	limitType byte
	let limitTypeKind = limitTypes[int(limitType)]
}

// --------------------------------------------------------------

Handshake0 = {
	h0 byte
	assert h0 == 3
}

Handshake1Verify = {
	let i = _diggestOffset
	let off = (_h[i] + _h[i+1] + _h[i+2] + _h[i+3]) % (764 - 32 - 4) + i + 4
	let data1 = _h[:off]
	let data2 = _h[off+32:]
	let digest = _h[off:off+32]
	let h = hmac.new(sha256.new, handshakeKey)
	do h.write(data1)
	do h.write(data2)
	let ok = bytes.equal(h.sum(nil), digest)
}

Handshake1 = {
	_h1 [1536]byte

	eval _h1 do {
		time	uint32be
		version uint32be
		global _h = _h1
		global _diggestOffset = 772
		_verify1 Handshake1Verify
		if _verify1.ok {
			let digest = _verify1.digest
			let msg = "Flash after 10.0.32.18"
		} else {
			global _diggestOffset = 8
			_verify2 Handshake1Verify
			if _verify2.ok {
				let digest = _verify2.digest
				let msg = "Flash before 10.0.32.18"
			}
		}
	}
}

Handshake2 = {
	_h2 [1536]byte
}

// --------------------------------------------------------------

AggregateItemHeader = {
	typeid   byte
	length   uint24be
	ts       uint24be
	streamid uint32be
}

AggregateItem = {
	header AggregateItemHeader
	_body  [header.length]byte
	back   uint32be
	assert back == header.length + 11
	eval _body do case header.typeid {
		18: AMF0
		20: AMF0_CMD
		15: AMF3
		17: AMF3_CMD
		8:  Audio
		9:  Video
		default: let body = _body
	}
}

AggregateMsg = {
	msgs *AggregateItem
}

// --------------------------------------------------------------

ChunkHeader = {
	_tag byte

	let format = (_tag >> 6) & 3
	assert format <= 3

	let csid = _tag & 0x3f
	if csid == 0 {
		_v byte
		let csid = _v + 0x40
	} elif csid == 1 {
		_v uint16le
		let csid = _v + 0x40
	}

	let _last = lastMsgs[csid]

	if format < 3 {
		ts uint24be
		if format < 2 {
			length uint24be
			typeid byte
			if format < 1 {
				streamid uint32le
			} else {
				let ts = ts + _last["ts"]
				let streamid = _last["streamid"]
			}
			let remain = 0
		} else {
			let ts = ts + _last["ts"]
			let length = _last["length"]
			let typeid = _last["typeid"]
			let streamid = _last["streamid"]
			let remain = _last["remain"]
		}
		if ts == 0xffffff {
			tsext uint32be
		}
	} else {
		let ts = _last["ts"]
		let length = _last["length"]
		let typeid = _last["typeid"]
		let streamid = _last["streamid"]
		let remain = _last["remain"]
		if ts == 0xffffff {
			tsext uint32be
		}
	}

	if remain == 0 {
		let remain = length
		let _body = bytes.buffer()
	} else {
		let _body = _last["body"]
	}
}

Chunk = {
	header ChunkHeader

	let _length = chunksize
	if header.remain < _length {
		let _length = header.remain
	}

	let _header = {
		"ts":	    header.ts,
		"length":   header.length,
		"typeid":   header.typeid,
		"streamid": header.streamid,
		"remain":   header.remain - _length,
		"body":	    header._body,
	}
	do set(lastMsgs, header.csid, _header)

	data [_length]byte
	do header._body.write(data)
	if _length > 16 {
		let data = data[:16]
	}

	if _header.remain == 0 {
		let _body = header._body.bytes()
		if header.csid == 2 && header.streamid == 0 {
			eval _body do case header.typeid {
				1: SetChunkSize
				2: Abort
				4: UserControl
				5: AckWinsize
				6: SetPeerBandwidth
				default: let body = _body
			}
		} else {
			eval _body do case header.typeid {
				18: AMF0
				20: AMF0_CMD
				15: AMF3
				17: AMF3_CMD
				22: AggregateMsg
				8:  Audio
				9:  Video
				default: let body = _body
			}
		}
	} elif RAWDATA == 0 {
		return undefined
	}
}

doc = init Handshake0 Handshake1 Handshake2 dump *(Chunk dump)

// --------------------------------------------------------------
