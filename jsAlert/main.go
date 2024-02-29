package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
	"unicode/utf8"
)

var (
  client      *http.Client
  transport   *http.Transport
  wg          sync.WaitGroup
  concurrency = 20
)

func main() {
  regexs := regexp.MustCompile(`(?i)((access_key|access_token|admin_pass|admin_user|algolia_admin_key|algolia_api_key|alias_pass|alicloud_access_key|amazon_secret_access_key|amazonaws|ansible_vault_pass|aos_key|api_key|api_key_secret|api_key_sid|api_secret|api.googlemaps AIza|apidocs|apikey|apiSecret|app_debug|app_id|app_key|app_log_level|app_secret|appkey|appkeysecret|application_key|appsecret|appspot|auth_token|authorizationTok|authsecret|aws_access_|aws_bucket|aws_key|aws_secret|aws_secret_key|aws_token|AWSSecretKey|b2_app_key|bashrc pass|bintray_apikey|bintray_gpg_pass|bintray_key|bintraykey|bluemix_api_key|bluemix_pass|browserstack_access_key|bucket_pass|bucketeer_aws_access_key_id|bucketeer_aws_secret_access_key|built_branch_deploy_key|bx_pass|cache_driver|cache_s3_secret|cattle_access|cattle_secret|certificate_pass|ci_deploy_pass|client_sec|client_zpk_secr|clojars_pass|cloud_api_|cloud_watch_aws_access_key|cloudant_pass|cloudflare_api_key|cloudflare_auth_key|cloudinary_api_secret|cloudinary_name|codecov_token|config|conn.login|connectionstring|consumer_key|consumer_secret|credentials|cypress_record_key|database_pass|database_schema_test|datadog_api_key|datadog_app_key|db_pass|db_server|db_username|dbpasswd|dbpass|dbuser|deploy_pass|digitalocean_ssh_key_body|digitalocean_ssh_key_ids|docker_hub_pass|docker_key|docker_pass|docker_pass|dockerhub_pass|dockerhubpass|dot-files|dotfiles|droplet_travis_pass|dynamoaccesskeyid|dynamosecretaccesskey|elastica_host|elastica_port|elasticsearch_pass|encryption_k|encryption_pass|env.heroku_api_key|env.sonatype_pass|eureka.awssecret)[a-z0-9_ .\-,]{0,25})(=|>|:=|\|\|:|<=|=>|:).{0,5}['\"]([0-9a-zA-Z\-_=]{8,64})['\"]`)

  sc := bufio.NewScanner(os.Stdin)

  client = &http.Client{
	Transport: &http.Transport{
	  MaxIdleConns:        concurrency,
	  MaxIdleConnsPerHost: concurrency,
	  MaxConnsPerHost:     concurrency,
	  TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	  },
	},
	Timeout: 5 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
	  return http.ErrUseLastResponse
	},
  }

  semaphore := make(chan bool, concurrency)

  for sc.Scan() {
	raw := sc.Text()
	wg.Add(1)
	semaphore <- true
	go func(raw string) {
	  defer wg.Done()
	  u := raw
	  resp, respBody, err := fetchURL(u)
	  if err != nil {
		return
	  }

	  if resp.StatusCode == 200 {
		submatchall := regexs.FindAllStringSubmatch(string(respBody), -1)
		if len(submatchall) > 0 {
		  for i := 0; i < len(submatchall); i++ {
			new := regexp.MustCompile(`.{0,10}` + submatchall[i][0] + `.{0,10}`)
			kk := new.FindStringSubmatch(string(respBody))
			fmt.Printf("%s: %s\n", u, kk)
		  }
		}
	  }
	}(raw)
	//here raw is the raw url and given as input to the function
	<-semaphore
  }
  wg.Wait()

  if sc.Err() != nil {
	fmt.Printf("error: %s\n", sc.Err())
  }
}

func fetchURL(u string) (*http.Response, string, error) {

  req, err := http.NewRequest("GET", u, nil)
  if err != nil {
	return nil, "", err
  }

  resp, err := client.Do(req)
  if err != nil {
	return nil, "", err
  }

  defer resp.Body.Close()

  respbody, err := io.ReadAll(resp.Body)
  resp.ContentLength = int64(utf8.RuneCountInString(string(respbody)))

  return resp, string(respbody), err
}
