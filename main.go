package main

import ( 
    "log"
    "net/http"
    "os"
    "github.com/samchelini/dns-manager/dns"
    "github.com/samchelini/dns-manager/jsend"
    "encoding/json"
    "fmt"
)

// global variables
var (
    port = "8080"   // default port
    tsig dns.TSIG   // tsig data
)

type Response[T any] struct {
    Resources   []T     `json:"resources"`
    Error       *string `json:"error"`  
}

type ResponseV2[T any] struct {
    Status      string  `json:"status"`
    Data        T       `json:"data"`
    Message     *string `json:"message,omitempty"`
    Code        *int    `json:"code,omitempty"`
}

// get all records from a zone
func getRecords(w http.ResponseWriter, r *http.Request) {
    // set headers and get zone from path
    w.Header().Set("Content-Type", "application/json")
    response := Response[dns.Record]{}
    zone := r.PathValue("zone")
    log.Printf("zone: %s", zone)

    // build and send query
    log.Println("building message...")
    query, err := dns.NewAxfrQuery(zone)
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

    // return response
    json.NewEncoder(w).Encode(response)
}

func sendResponse(writer http.ResponseWriter, response *jsend.Response) {
    writer.Header().Set("Content-Type", "application/json")
    writer.WriteHeader(response.HttpCode)
    json.NewEncoder(writer).Encode(response)
}

// get all records from a zone (v2)
func getRecordsV2(w http.ResponseWriter, r *http.Request) {
    // set headers and get zone from path
    w.Header().Set("Content-Type", "application/json")
    response := ResponseV2[[]dns.Record]{}
    zone := r.PathValue("zone")
    log.Printf("zone: %s", zone)

    // build and send query
    log.Println("building message...")
    query, err := dns.NewAxfrQueryV2(zone)
    if err != nil {
        sendResponse(w, err)
        return
    }
    response.Status = "success"
    answer, err := dns.SendQueryV2(query, os.Getenv("DNS_SERVER"))
    if err != nil {
        sendResponse(w, err)
        return
    }
    records, err := dns.GetAllRecordsV2(answer)
    if err != nil {
        sendResponse(w, err)
        return
    }
    response.Data = records
    w.WriteHeader(http.StatusOK)

    // return response
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

func parseEnv() error {
    // get env vars
    p := os.Getenv("PORT")
    dnsServer := os.Getenv("DNS_SERVER")
    tsigFile := os.Getenv("TSIG_FILE")

    // store any missing required env vars here
    missing := make([]string, 0)

    // check required env vars
    if dnsServer == "" {
        missing = append(missing, "DNS_SERVER")
    }
    if tsigFile == "" {
        missing = append(missing, "TSIG_FILE")
    }
    if len(missing) != 0 {
        missingString, _ := json.Marshal(missing)
        return fmt.Errorf("required env vars are missing: %s", missingString)
    }

    // try to open TSIG file
    _, err := os.Open(tsigFile)
    if err != nil {
        return fmt.Errorf("error opening TSIG_FILE: %s", err)
    }

    // try to read TSIG file
    tsigData, err := os.ReadFile(tsigFile)
    if err != nil {
        return fmt.Errorf("error reading TSIG_FILE: %s", err)
    }

    // try to unmarshal TSIG data to TSIG object
    err = json.Unmarshal(tsigData, &tsig)
    if err != nil {
        return fmt.Errorf("error unmarshalling TSIG_FILE: %s", err)
    }

    // check PORT
    if p == "" {
        log.Printf("PORT env var is not set, using default port %s", port)
    } else {
        log.Printf("PORT env var is set to: %s", p)
        port = p
    }

    return nil
}

func main() {
    // parse environment variables
    err := parseEnv()
    if err != nil {
        log.Fatalf("error parsing env vars: %s", err)
    }

    http.HandleFunc("GET /api/v1/records/{zone}", getRecords)
    http.HandleFunc("GET /api/v2/records/{zone}", getRecordsV2)
    http.HandleFunc("POST /api/v1/records/{zone}", updateRecord)
    http.HandleFunc("DELETE /api/v1/records/{zone}", updateRecord)
    log.Printf("listening on port %s ...", port)
    log.Fatal(http.ListenAndServe(":" + port, nil))
}
