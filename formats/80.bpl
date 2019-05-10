message = {
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

	} else {

		let _resp, _err = http.readResponse(BPL_IN, nil)
		assert _err == nil

		let status = _resp.status
		let statusCode = _resp.statusCode
		let header = _resp.header

		let _b, _err = ioutil.readAll(_resp.body)
		assert _err == nil

		let bodyLength = len(_b)
	}
}

doc = *(message dump)
