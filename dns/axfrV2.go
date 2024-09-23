package dns

import (
    "log"
    "encoding/binary"
    "golang.org/x/net/dns/dnsmessage"
    "github.com/samchelini/dns-manager/jsend"
    "net/http"
)

func NewAxfrQueryV2(domain string) ([]byte, *jsend.Response) {
    buf := make([]byte, 0)
    b := dnsmessage.NewBuilder(buf, dnsmessage.Header{
        ID: binary.BigEndian.Uint16(generateId()), 
        Response: false, 
        Authoritative: false,
    })
    b.EnableCompression()

    err := b.StartQuestions()
    if err != nil {
        log.Printf("error starting questions: %s", err)
        return nil, jsend.Error(domain, err.Error(), nil, http.StatusInternalServerError)
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
        return nil, jsend.Fail(domain, err.Error(), nil, http.StatusBadRequest)
    }

    query, err := b.Finish()
    if err != nil {
        log.Printf("error building message: %s", err)
        return nil, jsend.Error(domain, err.Error(), nil, http.StatusInternalServerError)
    }

    return query, nil
}

