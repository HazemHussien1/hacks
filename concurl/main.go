package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {

  var c int
  flag.IntVar(&c, "c", 10, "concurrency")

  flag.Parse()

  client := &http.Client{}
  client.Timeout = time.Second * 15

  sc := bufio.NewScanner(os.Stdin)

  var wg sync.WaitGroup
  urls := make(chan string)
  for i := 0; i < c; i++ {
	wg.Add(1)
	go func() {
	  for url := range urls {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
		  continue
		}
		resp, err := client.Do(req)
		if err != nil {
		  continue
		}

		defer resp.Body.Close()
		d, err := io.ReadAll(resp.Body)
		if err != nil {
		  continue
		}
		fmt.Printf("%s\n", string(d))
	  }
	  wg.Done()
	}()
  }

  for sc.Scan() {
	url := sc.Text()
	urls <- url
  }
  close(urls)
  wg.Wait()
}
