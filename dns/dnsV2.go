package dns

import (
    "log"
    "net"
    "encoding/binary"
    "golang.org/x/net/dns/dnsmessage"
    "github.com/samchelini/dns-manager/jsend"
    "net/http"
    "time"
)

const RCodeNotAuthorized dnsmessage.RCode = 9

var rCodeError = map[dnsmessage.RCode]string{
    dnsmessage.RCodeFormatError: "format error: the name server was unable to interpret the query.",
	dnsmessage.RCodeServerFailure: "server failure: the name server was unable to process this query due to a problem with the name server.",
	dnsmessage.RCodeNameError: "name error: the domain name referenced in the query does not exist.",
	dnsmessage.RCodeNotImplemented: "not implemented: the name server does not support the requested kind of query.",
	dnsmessage.RCodeRefused: "refused: the name server refuses to perform the specified operation for policy reasons.",
    RCodeNotAuthorized: "not authorized: server not authoritative for zone.",
}

// send query to nameserver and return answer (v2)
func SendQueryV2(query []byte, nameserver string) ([]byte, *jsend.Response) {
    length := make([]byte, 2)
    binary.BigEndian.PutUint16(length, uint16(len(query)))

    log.Printf("length: % x", length)
    query = append(length, query...)
    log.Printf("query: % x ", query)

    // send request
    log.Println("sending request...")
    conn, err := net.DialTimeout("tcp", nameserver, 5 * time.Second)
    if err != nil {
        log.Printf("error creating connection: %s", err)
        return nil, jsend.Error(nameserver, err.Error(), nil, http.StatusInternalServerError)
    }
    _, err = conn.Write(query)
    if err != nil {
        log.Printf("error sending query: %s", err)
        return nil, jsend.Error(nameserver, err.Error(), nil, http.StatusInternalServerError)
    }

    // receive answer
    answerLenBytes := make([]byte, 2)
    conn.Read(answerLenBytes)
    answerLen := binary.BigEndian.Uint16(answerLenBytes)
    log.Printf("answerLenBytes: % x ", answerLenBytes)
    log.Printf("answerLen: %d", answerLen)
    answer := make([]byte, answerLen)
    conn.Read(answer)
    log.Printf("answer: % x ", answer)
    log.Printf("length: % d", len(answer))

    return answer, nil
}

// get all records from answer (v2)
func GetAllRecordsV2(answer []byte) ([]Record, *jsend.Response) {
    // parse answer
    log.Println("starting parser...")
    var p dnsmessage.Parser
    if _, err := p.Start(answer); err != nil {
        log.Fatal(err)
    }
    // parse header
    header, err := p.Start(answer)
    if err != nil {
        log.Printf("error parsing header: %s", err)
        return nil, jsend.Error(nil, err.Error(), nil, http.StatusInternalServerError)
    }
    rCode := int(header.RCode)
    if rCode != 0 {
        log.Printf("dns error: rcode: %d: %s", rCode, rCodeError[header.RCode])
        return nil, jsend.Error(nil, rCodeError[header.RCode], &rCode, http.StatusInternalServerError)
    }
    err = p.SkipAllQuestions()
    if err != nil {
        log.Printf("error skipping questions: %s", err)
        return nil, jsend.Error(nil, err.Error(), nil, http.StatusInternalServerError)
    }

    log.Println("parsing answers...")
    var records []Record
    for {
		h, err := p.AnswerHeader()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
			log.Fatalf("error parsing answer: %s", err)
		}

        rec := Record {
            Name: h.Name.String(),
            Type: h.Type.String(),
            Class: h.Class.String(),
            TTL: h.TTL,
            Data: make(map[string]interface{}),
        }
        
		switch h.Type {
		case dnsmessage.TypeA:
			r, err := p.AResource()
			if err != nil {
				log.Fatalf("error parsing A Record: %s", err)
			}
			rec.Data["address"] = net.IP(r.A[:]).To4()
            records = append(records, rec)
        case dnsmessage.TypeSOA:
            r, err := p.SOAResource()
            if err != nil {
                log.Fatalf("error parsing SOA Record: %s", err)
            }
            rec.Data["ns"] = r.NS.String()
            rec.Data["mBox"] = r.MBox.String()
            rec.Data["serial"] = r.Serial
            rec.Data["refresh"] = r.Refresh
            rec.Data["retry"] = r.Retry
            rec.Data["expire"] = r.Expire
            rec.Data["minTtl"] = r.MinTTL
            records = append(records, rec)
        case dnsmessage.TypeNS:
            r, err := p.NSResource()
            if err != nil {
                log.Fatalf("error parsing NS Record: %s", err)
            }
            rec.Data["ns"] = r.NS.String()
            records = append(records, rec)
        case dnsmessage.TypePTR:
            r, err := p.PTRResource()
            if err != nil {
                log.Fatalf("error parsing PTR Record: %s", err)
            }
            rec.Data["ptr"] = r.PTR.String()
            records = append(records, rec)
        default:
            records = append(records, rec)
            p.SkipAnswer()
		}
	}

    return records, nil
}
