package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"../models"

	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/logs"
)

var client *http.Client

func init() {
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse("socks5://127.0.0.1:1080")
	}

	transport := &http.Transport{Proxy: proxy}

	client = &http.Client{Transport: transport}

}

type RespJson struct {
	Err      bool       `json:"error"`
	UrlsInfo []UrlsInfo `json:"body"`
}
type UrlsInfo struct {
	Urls Allurl `json:"urls"`
}
type Allurl struct {
	Origin  string `json:"original"`
	Regular string `json:"regular"`
}

//获取每日排名
func GetRankList() (*[]models.PicInfo, error) {

	geturl := "https://www.pixiv.net/ranking.php?mode=daily&content=illust&date=" + models.GetToday()
	request, err := http.NewRequest("GET", geturl, nil)
	if err != nil {
		logs.Error("url is err ", err)
		return nil, err
	}
	request.Header.Add("accept-language", "zh-CN,zh;q=0.9")
	resp, err := client.Do(request)
	if err != nil {
		logs.Error("无法获取这一天的数据", err)
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logs.Error(err)
		return nil, err
	}

	wrong := doc.Find(".error-unit")
	if exist := wrong.Text(); exist != "" {
		err := fmt.Errorf("昨天排行未出炉")
		logs.Error(err)
		return nil, err
	}

	list := doc.Find(".ranking-items")

	//图片id列表
	var pics []models.PicInfo
	list.Find("section").Each(func(i int, s *goquery.Selection) {
		//图片id
		id, _ := s.Attr("data-id")
		title, _ := s.Attr("data-title")
		pics = append(pics, models.PicInfo{
			ID:    id,
			Title: title,
		})

	})
	return &pics, nil
}

//通过图片id获取图片url
func GetPicURL(pid string) (origin *[]string, regular *[]string, err error) {
	origin = new([]string)
	regular = new([]string)
	detailURL := "https://www.pixiv.net/ajax/illust/" + pid + "/pages"
	resp, err := client.Get(detailURL)
	if err != nil {
		logs.Error(err)
		return nil, nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error(err)
		return nil, nil, err
	}
	var ResInfo RespJson
	err = json.Unmarshal(body, &ResInfo)
	if err != nil {
		logs.Error(err)
		return nil, nil, err
	}
	if ResInfo.Err {
		return nil, nil, fmt.Errorf("WRONG ANS OF PIXIV API")
	}

	for _, v := range ResInfo.UrlsInfo {
		*origin = append(*origin, v.Urls.Origin)
		*regular = append(*regular, v.Urls.Regular)
	}

	return
}

func DownLoadPic(id string, finURL string, savepath string) (err error) {
	request, err := http.NewRequest("GET", finURL, nil)
	if err != nil {
		logs.Error("url is err ", err)
		return
	}
	detailURL := "https://www.pixiv.net/member_illust.php?mode=medium&illust_id=" + id
	request.Header.Add("Referer", detailURL)

	//发送请求获取结果
	resp, err := client.Do(request)
	if err != nil {
		logs.Error(err)
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error(err)
		return
	}
	f, err := os.Create(savepath)
	defer f.Close()
	if err != nil {
		logs.Error(err)
		return
	}
	_, err = f.Write(body)
	if err != nil {
		logs.Error(err)
		return
	}

	return nil
}
