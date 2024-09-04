package main

import ( 
    "log"
    "net/http"
    "os"
    "github.com/samchelini/dns-manager/dns"
    "encoding/json"
)

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


func main() {
    if os.Getenv("DNS_SERVER") == "" {
        log.Fatal("error: DNS_SERVER env var not set")
    }

    http.HandleFunc("/api/v1/records", getRecords)
    log.Println("listening on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

