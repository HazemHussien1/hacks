package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func main() {

  var outputDir string
  // the default is the current directory just like wget
  flag.StringVar(&outputDir, "o", "./", "output directory")

  var headers headerArgs
  flag.Var(&headers, "H", "headers to add to each request")

  var proxy string
  flag.StringVar(&proxy, "x", "", "proxy server")

  var ignoreEmpty bool
  flag.BoolVar(&ignoreEmpty, "ie", false, "do not store empty files")

  flag.Parse()

  client := generateClient(proxy)
  prefix := outputDir
  if err := os.Mkdir(outputDir, os.ModePerm); err != nil {
	log.Fatal(err)
  }

  var wg sync.WaitGroup
  sc := bufio.NewScanner(os.Stdin)
  for sc.Scan() {
	rawURL := sc.Text()
	wg.Add(1)
	go func() {
	  defer wg.Done()

	  shouldSave := true

	  // create the request
	  _, err := url.ParseRequestURI(rawURL)
	  if err != nil {
		return
	  }

	  req, err := http.NewRequest("GET", rawURL, nil)
	  if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create request: %s\n", err)
		return
	  }

	  // add headers to the request
	  for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)

		if len(parts) != 2 {
		  continue
		}
		req.Header.Set(parts[0], parts[1])
	  }

	  // send the request
	  resp, err := client.Do(req)
	  if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return
	  }
	  defer resp.Body.Close()

	  // we want to read the body into a string or something like that so we can provide options to
	  // not save content based on a pattern or something like that
	  responseBody, err := io.ReadAll(resp.Body)
	  if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return
	  }

	  // sometimes we don't about the response at all if it's empty
	  if ignoreEmpty {
		if len(bytes.TrimSpace(responseBody)) > 0 {
		  shouldSave = true
		}
	  }

	  if !shouldSave {
		// fmt.Printf("%s %d\n", rawURL, resp.StatusCode)
		return
	  }

	  filenameToSave := getFileName(rawURL)
	  // write the response body to a file
	  if filenameToSave != "" {
		err = os.WriteFile(prefix + "/" + filenameToSave, responseBody, 0644)
		if err != nil {
		  fmt.Fprintf(os.Stderr, "failed to write file contents: %s\n", err)
		  return
		}
	  }
	}()
  }
  wg.Wait()
}

func generateClient(proxy string) *http.Client {

  tr := &http.Transport{
	MaxIdleConns:      30,
	IdleConnTimeout:   time.Second,
	TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	DialContext: (&net.Dialer{
	  Timeout:   time.Second * 10,
	  KeepAlive: time.Second,
	}).DialContext,
  }

  if proxy != "" {
	if p, err := url.Parse(proxy); err == nil {
	  tr.Proxy = http.ProxyURL(p)
	}
  }

  re := func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
  }

  return &http.Client{
	Transport:     tr,
	CheckRedirect: re,
	Timeout:       time.Second * 10,
  }
}

type headerArgs []string

func (h *headerArgs) Set(val string) error {
  *h = append(*h, val)
  return nil
}

func (h headerArgs) String() string {
  return strings.Join(h, ", ")
}

// https://www.domain.com/ -> https_www.domain.com
// https://www.domain.com/assets/ -> assets
// https://www.domain.com/assets/js/ -> js
// https://www.domain.com/assets/js/file.js -> file.js
func getFileName(rawURL string) string {
  uu,err := url.Parse(rawURL)
  if err != nil {
	fmt.Printf("error: %s\n",err)
	// here we extrace path here with regex
	// 1- remove ?.* from url
	var re = regexp.MustCompile(`\?.*`)
	rawURL = re.ReplaceAllString(rawURL, ``)

	// 2- remove #.* from url
	re = regexp.MustCompile(`#.*`)
	rawURL = re.ReplaceAllString(rawURL, ``)

	// 3- is url matches https://ww.d.com/?, replace 'https?://' with _ and return that
	re = regexp.MustCompile(`https?://[a-zA-Z-0-9\-_\.]+/?`)
	if re.MatchString(rawURL) {
	  re = regexp.MustCompile(`https?://`)
	  rawURL = re.ReplaceAllString(rawURL, `_`)
	  return rawURL
	}
	// 4- split url by "/" and return last part :)
	parts := strings.Split(rawURL,"/")
	return parts[len(parts)-1]
  }

  if uu.Path == "" || uu.Path == "/" {
	hostname := uu.Hostname()
	return uu.Scheme + "_" + hostname
  }

  path := uu.Path
  pathParts := strings.Split(path,"/")
  return pathParts[len(pathParts)-1]
}
