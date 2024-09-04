package dns
import (
    "log"
    "net"
    "crypto/rand"
    "encoding/binary"
    "golang.org/x/net/dns/dnsmessage"
    "time"
)

type Record struct {
    Name    string                  `json:"name"`
    Type    string                  `json:"type"`
    Class   string                  `json:"class"`
    TTL     uint32                  `json:"ttl"`
    Data    map[string]interface{}  `json:"data"`
}

// generate random 2 byte ID
func generateId() []byte {
    id := make([]byte, 2)
    _, err := rand.Read(id)
    if err != nil {
        log.Fatalf("error generating ID: %s", err)
    }
    return id
}

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

func SendQuery(query []byte, nameserver string) []byte {
    length := make([]byte, 2)
    binary.BigEndian.PutUint16(length, uint16(len(query)))

    log.Printf("length: % x", length)
    query = append(length, query...)
    log.Printf("query: % x ", query)

    // send request
    log.Println("sending request...")
    conn, err := net.DialTimeout("tcp", nameserver, 5 * time.Second)
    if err != nil {
        log.Fatalf("error creating connection: %s", err)
    }
    _, err = conn.Write(query)

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

    return answer
}

func GetRecords(answer []byte) []Record {
    // parse answer
    log.Println("parsing answer...")
    var p dnsmessage.Parser
    if _, err := p.Start(answer); err != nil {
        log.Fatal(err)
    }
    err := p.SkipAllQuestions()
    if err != nil {
        log.Fatalf("error skipping questions: %s", err)
    }

    log.Println("parsing answers...")
    //recs, err := p.AllAnswers()
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

    return records
}
