package dns

import (
    "log"
    "encoding/binary"
    "golang.org/x/net/dns/dnsmessage"
    "net"
    "strings"
    "time"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
)

// operation to use in udpate queries
type Op uint8
const (
    OpAdd       Op = 0
    OpDelete    Op = 1
)

// tsig object
type TSIG struct {
    Name        string  `json:"name"`
    Algorithm   string  `json:"algorithm"`
    Secret      string  `json:"secret"` 
}


// build an dynamic dns update query with tsig
func NewUpdateQuery(zone string, op Op, record *Record, tsig *TSIG) ([]byte, error) {
    buf := make([]byte, 0)
    id := generateId()

    // the update header (OpCode 5)
    b := dnsmessage.NewBuilder(buf, dnsmessage.Header{
        ID:         binary.BigEndian.Uint16(id),
        Response:   false,
        OpCode:     5,
    })

    // start zone section (same as question section)
    err := b.StartQuestions()
    if err != nil {
        log.Fatalf("error starting zones: %s", err)
    }
    err = b.Question(
        dnsmessage.Question{
            Name:   dnsmessage.MustNewName(zone),
            Type:   dnsmessage.TypeSOA,
            Class:  dnsmessage.ClassINET,
        },
    )
    if err != nil {
        log.Printf("error adding zone: %s", err)
        return nil, err
    }

    // start update section (same as authority section)
    err = b.StartAuthorities()
    if err != nil {
        log.Fatalf("error starting updates: %s", err)
    }

    // add records depending on the type of operation
    switch op {
    case OpDelete:
        deleteRecord(&b, record)
    case OpAdd:
        deleteRecord(&b, record)
        addRecord(&b, record)
    }

    // pass copy of the builder, 
    // since mac is generated based on the message before tsig record is added
    mac := GenerateMac(b, tsig)
    
    // start additional section (for tsig record)
    err = b.StartAdditionals()
    if err != nil {
        log.Fatalf("error starting additionals: %s", err)
    }

    // construct and add tsig record
    tsigHeader := dnsmessage.ResourceHeader {
        Name:   dnsmessage.MustNewName(tsig.Name),
        Type:   250, // tsig type
        Class:  dnsmessage.ClassANY,
        TTL:    0,
    }
    tsigResource := newTsigResource(tsig, mac, id)
    b.UnknownResource(tsigHeader, tsigResource)

    // finish building and return query
    query, err := b.Finish()
    if err != nil {
        log.Fatalf("error building message: %s", err)
    }
    return query, err
}

// adds a record to the builder question section
func addRecord(builder *dnsmessage.Builder, record *Record) {
    switch record.Type {
    case "TypeA":
        // convert address to [4]byte
        addr := [4]byte(net.ParseIP(record.Data["address"].(string)).To4())
        aResource := dnsmessage.AResource{A: addr}
        resourceHeader := dnsmessage.ResourceHeader {
            Name: dnsmessage.MustNewName(record.Name),
            Type: dnsmessage.TypeA,
            Class: dnsmessage.ClassINET,
            TTL: record.TTL,
        }
        builder.AResource(resourceHeader, aResource)
    }
    
}

// adds a delete record to the builder question section (class=any and ttl=0)
func deleteRecord(builder *dnsmessage.Builder, record *Record) {
    t := typeFromString(record.Type)
    resourceHeader := dnsmessage.ResourceHeader {
        Name: dnsmessage.MustNewName(record.Name),
        Type: t,
        Class: dnsmessage.ClassANY,
        TTL: 0,
    }
    resource := dnsmessage.UnknownResource{Type: t}
    builder.UnknownResource(resourceHeader, resource)
}

// returns a dnsmessage.Type for a string
func typeFromString(t string) (dnsmessage.Type) {
    switch t {
    case "TypeA":
        return dnsmessage.TypeA
    default:
        return dnsmessage.TypeALL
    }
}

// constructs and returns the tsig record data
func newTsigResource(tsig *TSIG, mac []byte, id []byte) (dnsmessage.UnknownResource) {
    tsigResource := dnsmessage.UnknownResource{Type: 250}
    data := make([]byte, 0)

    // append algorithm name
    tsigName := strings.Split(tsig.Algorithm, ".")
    for _, label := range tsigName {
        data = append(data, byte(len(label)))
        data = append(data, label...)
    }

    // append time signed
    time := time.Now().Unix()    
    log.Printf("time: %d", time)
    timeSigned := make([]byte, 8, 8)
    binary.BigEndian.PutUint64(timeSigned, uint64(time))
    log.Printf("timeSigned: % x", timeSigned)
    data = append(data, timeSigned[2:]...)

    // append fudge
    fudge := uint16(300)
    log.Printf("fudge: % x", fudge)
    data = binary.BigEndian.AppendUint16(data, fudge)
    log.Printf("TSIG data: % x", data)

    // append mac length and sum
    data = binary.BigEndian.AppendUint16(data, uint16(len(mac)))
    data = append(data, mac...)

    // append id
    data = append(data, id...)

    // append error
    data = binary.BigEndian.AppendUint16(data, uint16(0))

    // append other len
    data = binary.BigEndian.AppendUint16(data, uint16(0))
    log.Printf("TSIG data: % x", data)

    // set the tsig data and return the tsig resource
    tsigResource.Data = data
    return tsigResource
}

// generate mac from dns message (before TSIG RR is added)
func GenerateMac(builder dnsmessage.Builder, tsig *TSIG) []byte {
    msg, _ := builder.Finish()
    // add tsig data to dns message
    msg = append(msg, nameToWire(dnsmessage.MustNewName(tsig.Name).String())...) // name
    msg = binary.BigEndian.AppendUint16(msg, uint16(255)) // class
    msg = binary.BigEndian.AppendUint32(msg, uint32(0)) // ttl
    msg = append(msg, nameToWire(dnsmessage.MustNewName(tsig.Algorithm).String())...) // algorithm name

    // time signed
    time := time.Now().Unix()
    timeSigned := make([]byte, 8, 8)
    binary.BigEndian.PutUint64(timeSigned, uint64(time))
    msg = append(msg, timeSigned[2:]...)

    msg = binary.BigEndian.AppendUint16(msg, uint16(300)) // fudge
    msg = binary.BigEndian.AppendUint16(msg, uint16(0)) // error
    msg = binary.BigEndian.AppendUint16(msg, uint16(0)) // other len
    log.Printf("tsig digest: % x", msg)

    key, _ := base64.StdEncoding.DecodeString(tsig.Secret)
    log.Printf("key: % x", key)
    mac := hmac.New(sha256.New, key)
    mac.Write(msg)
	tsigMac := mac.Sum(nil)
    log.Printf("mac: % x", tsigMac)

    return tsigMac
}

// convert domain name to wire format
func nameToWire(name string) []byte {
    data := make([]byte, 0)
    labels := strings.Split(name, ".")
    for _, label := range labels {
        data = append(data, byte(len(label)))
        data = append(data, label...)
    }
    return data
}
