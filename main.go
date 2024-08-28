package main

import ( 
    "log"
    "net"
    "golang.org/x/net/dns/dnsmessage"
    "crypto/rand"
    "encoding/binary"
    "encoding/json"
    "github.com/gin-gonic/gin"
    "net/http"
)

type ARecord struct {
    Type string
    Name string
    Address string
}

func getRecords(c *gin.Context) {
    domain := c.Query("domain")
    log.Println("building message...")

    log.Println(domain)
    query := newAxfrQuery(domain)

    length := make([]byte, 2)
    binary.BigEndian.PutUint16(length, uint16(len(query)))

    log.Printf("length: % x", length)
    query = append(length, query...)
    log.Printf("query: % x ", query)

    // send request
    log.Println("sending request...")
    conn, err := net.Dial("tcp", "ns1.internal.chelini.io:53")
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

    // parse answer
    log.Println("parsing answer...")
    var p dnsmessage.Parser
	if _, err := p.Start(answer); err != nil {
		log.Fatal(err)
	}
    err = p.SkipAllQuestions()
    if err != nil {
        log.Fatalf("error skipping questions: %s", err)
    }

    log.Println("parsing answer headers...")
    var aRecords []ARecord
    for {
		h, err := p.AnswerHeader()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
            log.Fatalf("error parsing answer header: %s", err)
		}

        if h.Type == dnsmessage.TypeA {
            r, err := p.AResource()
			if err != nil {
                log.Fatalf("error parsing A resource: %s", err)
			}
            rType := h.Type.String()
            rName := h.Name.String()
            rAddr := net.IPv4(r.A[0], r.A[1], r.A[2], r.A[3]).String()
            rec := ARecord{
                Type: rType,
                Name: rName,
                Address: rAddr,
            }
            aRecords = append(aRecords, rec)
            log.Printf("type: %s", h.Type.String())
            log.Printf("name: %s", h.Name.String())
            log.Printf("addr: %s", net.IPv4(r.A[0], r.A[1], r.A[2], r.A[3]).To4()) 
        } else {
            p.SkipAnswer()
        }
	}
    j, _ := json.MarshalIndent(aRecords, "", "  ")
    log.Printf("printing json...\n%s", string(j))
    c.IndentedJSON(http.StatusOK, aRecords)
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

func newAxfrQuery(domain string) []byte {
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
        log.Fatalf("error adding question: %s", err)
    }

    query, err := b.Finish()
    if err != nil {
        log.Fatalf("error building message: %s", err)
    }

    return query
}

func main() {
    router := gin.Default()
    router.GET("/records", getRecords)

    router.Run("0.0.0.0:8080")
}

