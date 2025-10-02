package main

import (
	"flag"
	"fmt"
	"os"

	spamhaus "github.com/bryannolen/dnsbl/spamhaus"
)

var (
	rAddr = flag.String("resolver", "", "DNS resolver address, cannot use a public resolver like Google/CF/Quad9")
	qIP   = flag.String("ip", "", "IP to check")
)

func main() {
	flag.Parse()
	if *qIP == "" || *rAddr == "" {
		flag.Usage()
		return
	}
	res, err := spamhaus.QueryByIP(*qIP, *rAddr)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Printf("%v\n", res)
}
