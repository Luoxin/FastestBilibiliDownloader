package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"simple-golang-crawler/model"
	"simple-golang-crawler/tool"

	log "github.com/sirupsen/logrus"
)

var _startUrlTem = "https://api.bilibili.com/x/web-interface/view?aid=%d"

func GenVideoFetcher(video *model.Video) FetchFun {
	referer := fmt.Sprintf(_startUrlTem, video.ParCid.ParAid.Aid)
	for i := int64(1); i <= video.ParCid.Page; i++ {
		referer += fmt.Sprintf("/?p=%d", i)
	}

	return func(url string) (bytes []byte, err error) {
		<-_rateLimiter.C
		client := httpClientPool.GetClient()
		client.CheckRedirect = genCheckRedirectfun(referer)

		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalln(url, err)
			return nil, err
		}

		request.Header.Set("Referer", referer)

		resp, err := client.Do(request)
		if err != nil {
			log.Errorf("Fail to download the video %d,err is %s", video.ParCid.Cid, err)
			return nil, err
		}

		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			log.Fatalln("Fail to download the video %d,status code is %d", video.ParCid.Cid, resp.StatusCode)
			return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
		}
		defer resp.Body.Close()

		aidPath := tool.GetAidFileDownloadDir(video.ParCid.ParAid.Aid, video.ParCid.ParAid.Title)
		filename := fmt.Sprintf("%d_%d.flv", video.ParCid.Page, video.Order)
		file, err := os.Create(filepath.Join(aidPath, filename))
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
		defer file.Close()

		log.Println(video.ParCid.ParAid.Title + ":" + filename + " is downloading.")
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Printf("Failed to download video %d,err:%v", video.ParCid.Cid, err)
			return nil, err
		}
		log.Println(video.ParCid.ParAid.Title + ":" + filename + " has finished.")

		return nil, nil
	}
}

func genCheckRedirectfun(referer string) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		req.Header.Set("Referer", referer)
		return nil
	}
}
