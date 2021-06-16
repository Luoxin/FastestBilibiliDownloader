package fetcher

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var _rateLimiter = time.NewTicker(100 * time.Microsecond)

type FetchFun func(url string) ([]byte, error)

var fetcherClient = resty.New().
	SetTimeout(time.Second*5).
	SetRetryMaxWaitTime(time.Second*5).
	SetRetryWaitTime(time.Second).
	SetTimeout(time.Minute).
	SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36").
	SetLogger(log.New())

func DefaultFetcher(url string) ([]byte, error) {
	<-_rateLimiter.C
	client := fetcherClient.GetClient()
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("fetch err while request :%s,and the err is %s", url, err)
		return nil, err
	}
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0")

	resp, err := client.Do(request)
	if err != nil {
		log.Errorf("fetch err while request :%s,and the err is %s", url, err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	bodyReader := bufio.NewReader(resp.Body)

	e := determineEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())
	return ioutil.ReadAll(utf8Reader)
}

func determineEncoding(reader *bufio.Reader) encoding.Encoding {
	bytes, err := reader.Peek(1024)
	if err != nil {
		return unicode.UTF8
	}
	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}
