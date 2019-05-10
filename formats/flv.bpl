Chunk = {
	back       uint32be // 整个msg的长度（含flv头）
	typeid     uint8
	chunklen   uint24be
	ts         uint24be
	tsExtended uint8
	streamid   uint24be // always be 0?
	data       [chunklen]byte
}

FlvHeader = {
	tag [9]byte
	assert bytes.equal(tag, bytes.from([0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09]))
}

Flv = FlvHeader *(Chunk dump)

Message = {
	if BPL_DIRECTION == "REQ" {

		let _req, _err = http.readRequest(BPL_IN)
		assert _err == nil

		let method = _req.method
		let path = _req.URL.string()
		let host = _req.host
		let header = _req.header

		let _b, _err = ioutil.readAll(_req.body)
		assert _err == nil
		let body = string(_b)
		dump

	} else {

		let _resp, _err = http.readResponse(BPL_IN, nil)
		assert _err == nil

		let status = _resp.status
		let statusCode = _resp.statusCode
		let header = _resp.header
		dump

		eval _resp.body do Flv
	}
}

doc = *Message
