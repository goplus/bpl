const (
	fColorTable	        = 0x80
	fColorTableBitsMask = 7
)

ColorTable = {
	colortable [(1 << (1 + (fields&fColorTableBitsMask))) * 3]byte
}

Header = {
	tag             [6]char
	width           int16
	height          int16
	fields          uint8
	backgroundIndex uint8
	tmp	            uint8
	assert tag == "GIF87a" || tag == "GIF89a"
	if fields & fColorTable do ColorTable
}

sTrailer = nil

ImageHeader = {
	left   int16
	top	int16
	width  int16
	height int16
	fields byte
	if fields & fColorTable do ColorTable
}

ExtBlocks = {
	len byte
	if len {
		data [len]byte
		next ExtBlocks
	}
}

eComment = {
	blocks ExtBlocks
}

eApplication = {
	len	byte
	name   [len]char
	blocks ExtBlocks
}

eText = {
	text   [13]char
	blocks ExtBlocks
}

eGraphicControl = {
	unused1          byte
	flags            byte
	delayTime        int16
	transparentIndex byte
	unused2          byte
}

sExtension = {
	etag byte
	case etag {
		0x01: eText
		0xF9: eGraphicControl
		0xFE: eComment
		0xFF: eApplication
	}
}

sImage = {
	h        ImageHeader
	litWidth byte
	blocks   ExtBlocks
	assert litWidth >= 2 && litWidth <= 8
}

Record = {
	tag byte
	case tag {
		0x21: sExtension
		0x2C: sImage
		0x3B: sTrailer
	}
}

doc = Header dump *(Record dump)
