// -------------------------------------------------------------------------------------

listChunk = {
    ListType [4]char
    dump
    Chunks *chunk
}

fmtChunk = {
    Format uint16
    case Format {
        0x01: let FormatDesc = "PCM"
        0x03: let FormatDesc = "IEEE-Float"
        0x06: let FormatDesc = "ALAW"
        0x07: let FormatDesc = "MULAW"
        0x11: let FormatDesc = "ADPCM"
        0xfffe: let FormatDesc = "Extensible"
        default: let FormatDesc = "Unknown"
    }
    Channels uint16
    SamplesPerSec uint32
    BytesPerSec uint32
    BlockAlign uint16
    BitsPerSample uint16
    if Size > 16 {
        ExtraSize uint16
        ExtraData [ExtraSize]byte
        eval ExtraData {
            SamplesPerBlock uint16
        }
    }
    global nFormat = Format
    global nChannel = Channels
    global nBlockAlign = BlockAlign
    let Body = _body
    dump
}

factChunk = {
    DataFactSize uint32  // 数据转换为PCM格式后的大小
    let Body = _body
    dump
}

defaultChunk = {
    if Size < 128 {
        Data [Size]byte
    } else {
        Summary [128]byte
        skip Size-128
    }
    dump
}

// -------------------------------------------------------------------------------------

monoBlockHeader = {
    Sample0  uint16  // block 中第一个未压缩的采样值
    Index    byte
    Reserved byte
}

block = {
    Header [nChannel]monoBlockHeader
    Data   [nBlockAlign-nChannel*4]byte
    dump
}

dataChunk = case nFormat {
    0x11: *block
    default: defaultChunk
}

// -------------------------------------------------------------------------------------

chunk = {
    ID    [4]char
    Size  uint32
    _body [Size]byte
    eval _body do case ID {
        "LIST": listChunk
        "fmt ": fmtChunk
        "fact": factChunk
        "data": dataChunk
        default: defaultChunk
    }
}

riffHeader = {
    ID [4]char
    assert ID == "RIFF"

    Size   uint32
    Format [4]char
    dump
}

doc = riffHeader *chunk

// -------------------------------------------------------------------------------------
