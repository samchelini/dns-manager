package main

import ( 
    "log"
    "github.com/gin-gonic/gin"
    "net/http"
    "os"
    "github.com/samchelini/dns-manager/dns"
)


func getRecords(c *gin.Context) {
    domain := c.Query("domain")
    log.Println("building message...")

    log.Printf("domain: %s", domain)
    query, err := dns.NewAxfrQuery(domain)
    if err != nil {
        log.Println(err)
        return
    }

    answer := dns.SendQuery(query, os.Getenv("DNS_SERVER"))
    records := dns.GetRecords(answer)
    c.IndentedJSON(http.StatusOK, records)
}


func main() {
    if os.Getenv("DNS_SERVER") == "" {
        log.Fatal("error: DNS_SERVER env var not set")
    }
    router := gin.Default()
    router.GET("/records", getRecords)

    router.Run("0.0.0.0:8080")
}

