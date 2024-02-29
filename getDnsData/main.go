package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
)

func main() {

  var c int
  flag.IntVar(&c, "c", 10, "concurrency")
  flag.Parse()

  sc := bufio.NewScanner(os.Stdin)

  var wg sync.WaitGroup
  domains := make(chan string)

  for i := 0; i < c; i++ {
	wg.Add(1)
	go func() {
	  for domain := range domains {
		cname, err := net.LookupCNAME(domain)
		if err != nil {
		  fmt.Printf("error: %s\n",err)
		  cname = ""
		}

		ips, err := net.LookupIP(domain)

		if err != nil {
		  fmt.Printf("error: %s\n",err)
		  continue
		}

		for _, ip := range ips {
		  if ipv4 := ip.To4(); ipv4 != nil {
			fmt.Printf("%s\t\t\t%s\t\t\t%s\n",domain, ipv4, cname)
		  }
		}
	  }
	  wg.Done()
	}()
  }

  for sc.Scan() {
	domain := sc.Text()
	domains <- domain
  }
  close(domains)
  wg.Wait()

}
