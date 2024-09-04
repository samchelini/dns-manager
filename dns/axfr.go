package dns

import (
    "log"
    "encoding/binary"
    "golang.org/x/net/dns/dnsmessage"
)

func NewAxfrQuery(domain string) ([]byte, error) {
    buf := make([]byte, 0)
    b := dnsmessage.NewBuilder(buf, dnsmessage.Header{
        ID: binary.BigEndian.Uint16(generateId()), 
        Response: false, 
        Authoritative: false,
    })
    b.EnableCompression()

    err := b.StartQuestions()
    if err != nil {
        log.Fatalf("error starting questions: %s", err)
    }

    err = b.Question(
        dnsmessage.Question{
            Name: dnsmessage.MustNewName(domain), 
            Type: dnsmessage.TypeAXFR, 
            Class: dnsmessage.ClassINET,
        },
    )
    if err != nil {
        log.Printf("error adding question: %s", err)
        return nil, err
    }

    query, err := b.Finish()
    if err != nil {
        log.Fatalf("error building message: %s", err)
    }

    return query, err
}

