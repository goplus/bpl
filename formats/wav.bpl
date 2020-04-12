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
    }
    let Body = _body
    dump
}

factChunk = {
    Samples uint32
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

chunk = {
    ID    [4]char
    Size  uint32
    _body [Size]byte
    eval _body do case ID {
        "LIST": listChunk
        "fmt ": fmtChunk
        "fact": factChunk
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
