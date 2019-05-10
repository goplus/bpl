init = {
	global filterDir = BPL_FILTER["dir"]

	if filterDir != undefined {
		if filterDir != BPL_DIRECTION {
			do exit(0)
		}
	}

	global msgs = mkmap("int:var")
	global chunksize = 128
	global typeidKinds = {
		1:  "SetChunkSize",
		2:  "Abort",
		4:  "UserControl",
		5:  "AckWinsize",
		6:  "SetPeerBandwidth",
		8:  "Audio",
		9:  "Video",
		18: "AMF0",
		20: "AMF0_CMD",
		15: "AMF3",
		17: "AMF3_CMD",
	}
}

SetChunkSize = {
	size uint32be
	let chunksize = size
}

Abort = {
	csid uint32be
	let _last = msgs[csid]
	do set(_last, "remain", 0)
}

Handshake0 = {
	h0 [1]byte
}

Handshake1 = {
	h1 [1536]byte
}

Handshake2 = {
	h2 [1536]byte
}

ChunkHeader = {

	global b = bytes.buffer()

	_tag byte
	do b.writeByte(_tag)

	let format = (_tag >> 6) & 3
	assert format <= 3

	let csid = _tag & 0x3f
	if csid == 0 {
		_v byte
		do b.writeByte(_v)
		let csid = _v + 0x40
	} elif csid == 1 {
		_v2 [2]byte
		eval _v2 do {
			_v uint16le
			let csid = _v + 0x40
			do b.write(_v2)
		}
	}

	let _last = msgs[csid]

	if format < 3 {
		_v3 [3]byte
		eval _v3 do {
			ts uint24be
			do b.write(_v3)
		}
		if format < 2 {
			_v4 [4]byte
			eval _v4 do {
				length uint24be
				typeid byte
				do b.write(_v4)
			}
			if format < 1 {
				_v5 [4]byte
				eval _v5 do {
					streamid uint32le
					do b.write(_v5)
				}
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
	} else {
		let ts = _last["ts"]
		let length = _last["length"]
		let typeid = _last["typeid"]
		let streamid = _last["streamid"]
		let remain = _last["remain"]
	}

	let typeidKind = typeidKinds[int(typeid)]
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
		"ts": header.ts,
		"length": header.length,
		"typeid": header.typeid,
		"streamid": header.streamid,
		"remain": header.remain - _length,
		"body": header._body,
	}
	do set(msgs, header.csid, _header)

	_data [_length]byte
	do header._body.write(_data)
	do b.write(_data)
	let raw = b.bytes()

	if _header.remain == 0 {
		let _body = header._body.bytes()
		if header.csid == 2 && header.streamid == 0 {
			eval _body do case header.typeid {
				1: SetChunkSize
				2: Abort
				default: nil
			}
		}
	}
}

doc = init Handshake0 Handshake1 Handshake2 dump *(Chunk dump)
