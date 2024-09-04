package main

import ( 
    "log"
    "github.com/gin-gonic/gin"
    "net/http"
    "os"
    "github.com/samchelini/dns-manager/dns"
)

type Response[T any] struct {
    Resources   []T     `json:"resources"`
    Error       *string `json:"error"`  
}

func getRecords(c *gin.Context) {
    res := Response[dns.Record]{}
    domain := c.Query("domain")
    log.Printf("domain: %s", domain)

    log.Println("building message...")
    query, err := dns.NewAxfrQuery(domain)
    if err != nil {
        s := err.Error()
        res.Error = &s
        c.IndentedJSON(http.StatusBadRequest, res)
        return
    }

    answer := dns.SendQuery(query, os.Getenv("DNS_SERVER"))
    records := dns.GetRecords(answer)
    res.Resources = records
    c.IndentedJSON(http.StatusOK, res)
}


func main() {
    if os.Getenv("DNS_SERVER") == "" {
        log.Fatal("error: DNS_SERVER env var not set")
    }
    router := gin.Default()
    router.GET("/records", getRecords)

    router.Run("0.0.0.0:8080")
}

