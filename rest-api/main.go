package main

import (
    "log"
    "fmt"
    "net/http"
    "io/ioutil"
    "net"
    "os"
    "bufio"
    "os/exec"
    "bytes"
    "encoding/json"

    "dyndns/ipparser"
    "github.com/gorilla/mux"
)

var appConfig = &Config{}

func main() {
    appConfig.LoadConfig("/etc/dyndns.json")
    log.SetFlags(0)

    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/update", Update).Methods("GET")

    log.Println(fmt.Sprintf("Serving dyndns REST services on 0.0.0.0:8080..."))
    log.Fatal(http.ListenAndServe(":8080", router))
}

func Update(w http.ResponseWriter, r *http.Request) {
    response := BuildWebserviceResponseFromRequest(r, appConfig)

    if response.Success == false {
        json.NewEncoder(w).Encode(response)
        return
    }

    for _, domain := range response.Domains {
        ipExists := CheckIpFromDns(response.Address, fmt.Sprintf("%s.%s", domain, appConfig.Domain))

        if ipExists {
            response.Success = true
            response.Message = "Record exist already"

            log.Println(fmt.Sprintf("No update is needed for %s.%s IP address %s", domain, appConfig.Domain, response.Address))
            json.NewEncoder(w).Encode(response)
            return
        }

        result := UpdateRecord(domain, response.Address, response.AddrType)

        if result != "" {
            response.Success = false
            response.Message = result

            json.NewEncoder(w).Encode(response)
            return
        }
    }

    response.Success = true
    response.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", response.AddrType, response.Domain, response.Address)

    json.NewEncoder(w).Encode(response)
}

func CheckIpFromDns(ipaddr string, domain string) bool {
    var ipExists bool = false
    ips, err := net.LookupIP(domain)

    if err != nil {
        log.Println(fmt.Sprintf("No IP from DNS, set new IP to %s", ipaddr))
        ipExists = false
    }

    for _, ip := range ips {
        if ipparser.ValidIP4(ip.String()) && ip.String() == ipaddr {
            ipExists = true
            break
        } else if ipparser.ValidIP6(ip.String()) && ip.String() == ipaddr {
            ipExists = true
            break
        }
    }

    return ipExists
}

func UpdateRecord(domain string, ipaddr string, addrType string) string {
    log.Println(fmt.Sprintf("%s record update request: %s -> %s", addrType, domain, ipaddr))

    f, err := ioutil.TempFile(os.TempDir(), "dyndns")
    if err != nil {
        return err.Error()
    }

    defer os.Remove(f.Name())
    w := bufio.NewWriter(f)

    w.WriteString(fmt.Sprintf("server %s\n", appConfig.Server))
    w.WriteString(fmt.Sprintf("zone %s\n", appConfig.Zone))
    w.WriteString(fmt.Sprintf("update delete %s.%s %s\n", domain, appConfig.Domain, addrType))
    w.WriteString(fmt.Sprintf("update add %s.%s %v %s %s\n", domain, appConfig.Domain, appConfig.RecordTTL, addrType, ipaddr))
    w.WriteString("send\n")

    w.Flush()
    f.Close()

    cmd := exec.Command(appConfig.NsupdateBinary, f.Name())
    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr
    err = cmd.Run()
    if err != nil {
        return err.Error() + ": " + stderr.String()
    }

    return out.String()
}
