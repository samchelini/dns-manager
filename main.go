package main

import ( 
    "log"
    "net/http"
    "os"
    "github.com/samchelini/dns-manager/dns"
    "encoding/json"
)

var tsig dns.TSIG

type Response[T any] struct {
    Resources   []T     `json:"resources"`
    Error       *string `json:"error"`  
}

func getRecords(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    response := Response[dns.Record]{}
    domain := r.FormValue("domain")
    log.Printf("domain: %s", domain)

    log.Println("building message...")
    query, err := dns.NewAxfrQuery(domain)
    if err != nil {
        errString := err.Error()
        response.Error = &errString
        w.WriteHeader(http.StatusBadRequest)
    } else {
        answer := dns.SendQuery(query, os.Getenv("DNS_SERVER"))
        records := dns.GetAllRecords(answer)
        response.Resources = records
        w.WriteHeader(http.StatusOK)
    }

    json.NewEncoder(w).Encode(response)
}

// create or delete dns record in a zone
func updateRecord(w http.ResponseWriter, r *http.Request) {
    var rec dns.Record
    var response Response[dns.Record]
    
    // decode provided json record to record object
    err := json.NewDecoder(r.Body).Decode(&rec)
    if err != nil {
        log.Println("error decoding: %s", err)
        errString := err.Error()
        response.Error = &errString
        w.WriteHeader(http.StatusBadRequest)
    }

    log.Printf("zone: %s", r.PathValue("zone"))

    // create query based on method type
    var query []byte
    switch r.Method {
    case "POST":
        query, err = dns.NewUpdateQuery(r.PathValue("zone"), dns.OpAdd, &rec, &tsig)
    case "DELETE": 
        query, err = dns.NewUpdateQuery(r.PathValue("zone"), dns.OpDelete, &rec, &tsig)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    // send query and write response
    dns.SendQuery(query, os.Getenv("DNS_SERVER"))
    response.Resources = append(response.Resources, rec)
    json.NewEncoder(w).Encode(response)
}

func main() {
    // check env vars
    if (os.Getenv("DNS_SERVER") == "") {
        log.Fatal("error: DNS_SERVER env var not set")
    }
    if (os.Getenv("TSIG_FILE") == "") {
        log.Fatal("error: TSIG_FILE env var not set")
    } else {
        // try to open file
        _, err := os.Open(os.Getenv("TSIG_FILE"))
        if err != nil {
            log.Fatalf("error opening TSIG_FILE: %s", err)
        }
        // try to read file
        tsigData, err := os.ReadFile(os.Getenv("TSIG_FILE"))
        if err != nil {
            log.Fatalf("error reading TSIG_FILE: %s", err)
        }
        // try to unmarshal to TSIG object
        err = json.Unmarshal(tsigData, &tsig)
        if err != nil {
            log.Fatalf("error unmarshalling TSIG_FILE: %s", err)
        }
    }

    http.HandleFunc("GET /api/v1/records", getRecords)
    http.HandleFunc("POST /api/v1/records/{zone}", updateRecord)
    http.HandleFunc("DELETE /api/v1/records/{zone}", updateRecord)
    log.Println("listening on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

