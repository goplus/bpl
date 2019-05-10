//
// http://blog.sina.com.cn/s/blog_48f93b530100jz4b.html
// http://www.52rd.com/Blog/wqyuwss/559/
// http://blog.csdn.net/wutong_login/article/category/567011

box = {
	size  uint32be
	typ   [4]char
	if size == 1 {
		_largesize uint64be
		let size = _largesize - 8
	}

	if size == 0 {
		let _body = mkslice("byte", 0)
	} elif typ == "mdat" {
		skip size - 8
		let _body = mkslice("byte", 0)
	} else {
		_body [size - 8]byte
	}
}

boxtr = {
	let body = _body
}

fixed16be = {
	_v uint16be
	return float64(_v) / 0x100
}

fixed32be = {
	_v uint32be
	return float64(_v) / 0x10000
}

// --------------------------------------------------------------

avc1 = {
	body *byte
}

mp4a = {
	body *byte
}

uuid = {
	body *byte
}

stsdbox = box {
	eval _body do case typ {
		"avc1": avc1
		default: boxtr
	}
}

stsd = {
	let class = "Sample Description"
	version byte
	flags   uint24be
	nvals   uint32be
	vals    [nvals]stsdbox
}

// --------------------------------------------------------------

stts = {
	body *byte
}

stsz = {
	let class = "Sample Size"
	let bodyLength = len(_body)
}

stz2 = {
	body *byte
}

stsc = {
	// “stss”确定media中的关键帧。对于压缩媒体数据，关键帧是一系列压缩序列的开始帧，其解压缩时不依赖以前的帧，而后续帧的解压缩将依赖于这个关键帧。
	// “stss”可以非常紧凑的标记媒体内的随机存取点，它包含一个sample序号表，表内的每一项严格按照sample的序号排列，说明了媒体中的哪一个sample是关键帧。
	// 如果此表不存在，说明每一个sample都是一个关键帧，是一个随机存取点。
	//
	let class = "Sample To Chunk Box"
	let bodyLength = len(_body)
}

stco = {
	// ”stco”定义了每个thunk在媒体流中的位置。位置有两种可能，32位的和64位的，后者对非常大的电影很有用。
	// 在一个表中只会有一种可能，这个位置是在整个文件中的，而不是在任何box中的，这样做就可以直接在文件中找到媒体数据，而不用解释box。
	// 需要注意的是一旦前面的box有了任何改变，这张表都要重新建立，因为位置信息已经改变了。
	//
	let class = "Chunk Offset"
	let bodyLength = len(_body)
}

co64 = {
	let class = "Chunk Offset64"
	let bodyLength = len(_body)
}

ctts = {
	let class = "Composition Time To Sample"
	let bodyLength = len(_body)
}

stss = {
	// “stss”确定media中的关键帧。对于压缩媒体数据，关键帧是一系列压缩序列的开始帧，其解压缩时不依赖以前的帧，而后续帧的解压缩将依赖于这个关键帧。
	// “stss”可以非常紧凑的标记媒体内的随机存取点，它包含一个sample序号表，表内的每一项严格按照sample的序号排列，说明了媒体中的哪一个sample是关键帧。
	// 如果此表不存在，说明每一个sample都是一个关键帧，是一个随机存取点。
	//
	let class = "Sync Sample"
	let bodyLength = len(_body)
}

stbbox = box {
	eval _body do case typ {
		"stsd": stsd
		"stts": stts
		"stsz": stsz
		"stz2": stz2
		"stsc": stsc
		"stco": stco
		"co64": co64
		"ctts": ctts
		"stss": stss
		default: boxtr
	}
}

// Sample Table Box
//
stbl = {
	vals *stbbox
}

// --------------------------------------------------------------

dref = {
	body *byte
}

dibox = box {
	eval _body do case typ {
		"dref": dref
		default: boxtr
	}
}

dinf = {
	vals *dibox
}

// --------------------------------------------------------------

vmhd = {
	body *byte
}

smhd = {
	body *byte
}

hmhd = {
	body *byte
}

nmhd = {
	body *byte
}

mibox = box {
	eval _body do case typ {
		"vmhd": vmhd
		"smhd": smhd
		"hmhd": hmhd
		"nmhd": nmhd
		"dinf": dinf
		"stbl": stbl
		default: boxtr
	}
}

// Media Information Box
//
minf = {
	vals *mibox
}

// --------------------------------------------------------------

mdhd = {
	version byte
	flags   uint24be

	ctime uint32be  // 创建时间（相对于UTC时间1904-01-01零点的秒数）
	mtime uint32be  // 修改时间

	time_scale uint32be // 文件媒体在1秒时间内的刻度值，可以理解为1秒长度的时间单元数
	duration   uint32be // 该track的时间长度，用duration和time_scale值可以计算track时长

	language   uint16be
	predefined uint16be
}

// Handler Reference Box
//
hdlr = {
	version byte
	flags   uint24be

	predefined uint32be

	// 在media box中，该值为4个字符：
	// “vide”— video track
	// “soun”— audio track
	// “hint”— hint track
	//
	handler_typ [4]char

	reserved [12]byte
	name     cstring // track type name，以‘\0’结尾的字符串
}

mdbox = box {
	eval _body do case typ {
		"mdhd": mdhd
		"hdlr": hdlr
		"minf": minf
		default: boxtr
	}
}

// Media Box
//
mdia = {
	vals *mdbox
}

// --------------------------------------------------------------

elst = {
	body *byte
}

edtsbox = box {
	eval _body do case typ {
		"elst": elst
		default: boxtr
	}
}

edts = {
	vals *edtsbox
}

// --------------------------------------------------------------

// Track Header Box
//
tkhd = {
	version byte

	// 按位或操作结果值，预定义如下：
	// 0x000001 track_enabled，否则该track不被播放；
	// 0x000002 track_in_movie，表示该track在播放中被引用；
	// 0x000004 track_in_preview，表示该track在预览时被引用。
	// 一般该值为7，如果一个媒体所有track均未设置track_in_movie和track_in_preview，将被理解为所有track均设置了这两项；
	// 对于hint track，该值为0
	//
	flags uint24be

	ctime uint32be  // 创建时间（相对于UTC时间1904-01-01零点的秒数）
	mtime uint32be  // 修改时间

	track_id uint32be // track id号，不能重复且不能为0

	reserved  uint32be
	duration  uint32be
	reserved2 uint64be

	layer     uint16be  // 视频层，默认为0，值小的在上层
	alt_group uint16be  // alternate group: track分组信息，默认为0表示该track未与其他track有群组关系
	volume    fixed16be // [8.8] 格式，1.0（0x0100）表示最大音量。如果为音频track，该值有效，否则为0
	reserved3 uint16be

	matrix [36]byte  // 视频变换矩阵
	width  fixed32be // 宽
	height fixed32be // 高，均为 [16.16] 格式值，与sample描述中的实际画面大小比值，用于播放时的展示宽高
}

trkbox = box {
	eval _body do case typ {
		"tkhd": tkhd
		"mdia": mdia
		"edts": edts
		default: boxtr
	}
}

// Track Box
//
trak = {
	vals *trkbox
}

// --------------------------------------------------------------

// Movie Header Box
//
mvhd = {
	version    byte
	flags      uint24be
	ctime      uint32be  // 创建时间（相对于UTC时间1904-01-01零点的秒数）
	mtime      uint32be  // 修改时间
	time_scale uint32be  // 文件媒体在1秒时间内的刻度值，可以理解为1秒长度的时间单元数
	duration   uint32be  // 该track的时间长度，用duration和time_scale值可以计算track时长
	rate       fixed32be // 推荐播放速率，高16位和低16位分别为小数点整数部分和小数部分，即[16.16] 格式，该值为1.0（0x00010000）表示正常前向播放
	volume     fixed16be // 与rate类似，[8.8] 格式，1.0（0x0100）表示最大音量
	reserved   [10]byte
	matrix     [36]byte  // 视频变换矩阵
	predefined [24]byte

	// 下一个track使用的id号
	//
	next_track_id uint32be
}

iods = {
	let class = "Initial Object Descriptor"
	version byte
	flags   uint24be
	unknown *byte
}

movbox = box {
	eval _body do case typ {
		"mvhd": mvhd
		"iods": iods
		"trak": trak
		default: boxtr
	}
	dump
}

// Movie Box
//
moov = dump *movbox

// --------------------------------------------------------------

// File Type Box
//
ftyp = {
	major_brand       [4]char
	minor_version     uint32be
	compatible_brands *[4]char
	dump
}

free = {
	let bodyLength = len(_body)
	dump
}

mdat = {
	dump
}

gblbox = box {
	//
	// 注意：moov 可能太大了，如果等解析完再 dump 不太友好
	//
	eval _body do case typ {
		"ftyp": ftyp
		"moov": moov
		"free": free
		"mdat": mdat
		default: boxtr dump
	}
}

doc = *gblbox

// --------------------------------------------------------------
