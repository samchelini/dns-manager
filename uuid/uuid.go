package uuid

import (
    "math/rand/v2"
    "fmt"
)

type UUID struct {
    timeLow             uint32
    timeMid             uint16
    timeHighAndVersion  uint16
    clockSeqAndReserved uint8
    clockSeqLow         uint8
    node                uint64
}

// return string in format defined in rfc4122
func (u UUID) String() string {
    return fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%012x", u.timeLow, u.timeMid, u.timeHighAndVersion, u.clockSeqAndReserved, u.clockSeqLow, u.node >> 16)
}

// generate a UUIDv4 (random)
func V4() string {
    var uuid UUID
    var ver uint16 = 0b0100 << 12   // version 4
    var reserved uint8 = 0b10 << 6  // rfc4122 reserved bits
    uuid.timeLow = rand.Uint32()
    uuid.timeMid = uint16(rand.Uint32() >> 16)
    uuid.timeHighAndVersion = uint16(rand.Uint32() >> 20) | ver
    uuid.clockSeqAndReserved = uint8(rand.Uint32() >> 26) | reserved
    uuid.clockSeqLow = uint8(rand.Uint32() >> 24)
    uuid.node = rand.Uint64()
    return uuid.String()
}
