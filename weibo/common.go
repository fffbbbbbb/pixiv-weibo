package weibo

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/logs"
)

var (
	username = "812129358@qq.com"
	password = "123"
)

func SignIn() (*http.Client, string, error) {
	client := &http.Client{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client.Jar = jar
	geturl := "https://login.sina.com.cn/sso/prelogin.php?entry=weibo&callback=sinaSSOController.preloginCallBack&su=" + base64.URLEncoding.EncodeToString([]byte(username)) + "&rsakt=mod&checkpin=1&client=ssologin.js(v1.4.19)&_=" + fmt.Sprintf("%v", time.Now().UnixNano()/1e6)
	request, err := http.NewRequest("GET", geturl, nil)
	if err != nil {
		logs.Error("url is err ", err)
		return nil, "", nil
	}
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	request.Header.Add("accept-language", "zh-CN,zh;q=0.9")
	resp, err := client.Do(request)
	if err != nil {
		logs.Error(err)
		return nil, "", nil
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	index := strings.Index(string(body), "(")
	var tmp []byte
	for i := index + 1; i < len(body); i++ {
		tmp = append(tmp, body[i])
	}
	index = strings.Index(string(tmp), ")")
	var tmp2 []byte
	for i := 0; i < index; i++ {
		tmp2 = append(tmp2, tmp[i])
	}
	var mapResult map[string]interface{}
	err = json.Unmarshal(tmp2, &mapResult)
	if err != nil {
		fmt.Println("JsonToMapDemo err: ", err)
	}

	nonce := mapResult["nonce"].(string)
	pubkey := mapResult["pubkey"].(string)
	rsakv := mapResult["rsakv"].(string)
	servertime := strconv.FormatFloat(mapResult["servertime"].(float64), 'f', -1, 64)

	DataURLVal := url.Values{}
	DataURLVal.Add("entry", "weibo")
	DataURLVal.Add("gateway", "1")
	DataURLVal.Add("from", "")
	DataURLVal.Add("savestate", "0")
	DataURLVal.Add("qrcode_flag", "false")
	DataURLVal.Add("useticket", "1")
	DataURLVal.Add("pagerefer", "")
	DataURLVal.Add("vsnf", "1")
	DataURLVal.Add("su", base64.URLEncoding.EncodeToString([]byte(username)))
	DataURLVal.Add("service", "miniblog")
	DataURLVal.Add("servertime", servertime)
	DataURLVal.Add("nonce", nonce)
	DataURLVal.Add("pwencode", "rsa2")
	DataURLVal.Add("rsakv", rsakv)
	DataURLVal.Add("sp", encryptPassword(pubkey, string(servertime), nonce, password))
	DataURLVal.Add("sr", "1920*1080")
	DataURLVal.Add("encoding", "UTF-8")
	DataURLVal.Add("prelt", "201")
	DataURLVal.Add("url", "https://weibo.com/ajaxlogin.php?framelogin=1&callback=parent.sinaSSOController.feedBackUrlCallBack")
	DataURLVal.Add("returntype", "META")

	request, err = http.NewRequest("POST", "https://login.sina.com.cn/sso/login.php?client=ssologin.js(v1.4.19)", strings.NewReader(DataURLVal.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	// request.Header.Add("Cookie", firstCookie[0])
	request.Header.Add("Referer", "https://weibo.com/")
	resp, err = client.Do(request)
	if err != nil {
		logs.Error(err)
		return nil, "", nil
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil, "", nil
	}
	threeUrl := doc.Find("script").Text()
	index = strings.Index(threeUrl, `"`)
	threeUrl = threeUrl[index+1 : len(threeUrl)]
	index = strings.Index(threeUrl, `"`)
	threeUrl = threeUrl[0 : index-1]

	request, err = http.NewRequest("GET", threeUrl, nil)
	if err != nil {
		logs.Error("url is err ", err)
		return nil, "", nil
	}
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	resp, err = client.Do(request)
	if err != nil {
		logs.Error(err)
		return nil, "", nil
	}
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)
	index = strings.Index(string(body), `[`)
	ticket := string(body)[index+1 : len(string(body))]
	index = strings.Index(ticket, `]`)
	ticket = ticket[0:index]

	urls := strings.Replace(ticket, `"`, "", -1) //替换tab为空格
	urls = strings.Replace(urls, `\`, "", -1)    //替换tab为空格
	urlarr := strings.Split(urls, ",")

	fourUrl := urlarr[0]
	request, err = http.NewRequest("GET", fourUrl, nil)
	if err != nil {
		logs.Error("url is err ", err)
		return nil, "", nil
	}
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	resp, err = client.Do(request)
	if err != nil {
		logs.Error(err)
		return nil, "", nil
	}
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)

	index = strings.Index(string(body), `":"`)
	uniqueid := string(body)[index+3 : len(string(body))]
	index = strings.Index(uniqueid, `"`)
	uniqueid = uniqueid[0:index]
	fiveUrl := "https://weibo.com/u/" + uniqueid + "/home?wvr=5&lf=reg"
	request, err = http.NewRequest("GET", fiveUrl, nil)
	if err != nil {
		logs.Error("url is err ", err)
		return nil, "", nil
	}
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	resp, err = client.Do(request)
	if err != nil {
		logs.Error(err)
		return nil, "", nil
	}
	defer resp.Body.Close()

	return client, uniqueid, nil
}

func UploadPic(client *http.Client, uniqueid, path string) (string, error) {

	ff, _ := os.Open(path)
	defer ff.Close()
	// sourcebuffer := make([]byte, 500000)
	n, _ := ioutil.ReadAll(ff)
	sourcestring := base64.StdEncoding.EncodeToString(n)
	uploadURL := `https://picupload.weibo.com/interface/pic_upload.php?cb=https%3A%2F%2Fweibo.com%2Faj%2Fstatic%2Fupimgback.html%3F_wv%3D5%26callback%3DSTK_ijax_` + fmt.Sprintf("%v", time.Now().UnixNano()/1e6) + `&mime=image%2Fjpeg&data=base64&url=weibo.com%2Fu%2F` + uniqueid + `&markpos=1&logo=1&nick=%40ff--FZ&marks=0&app=miniblog&s=rdxt&pri=null&file_source=1`
	uploadData := url.Values{}
	uploadData.Add("b64_data", sourcestring)
	request, err := http.NewRequest("POST", uploadURL, strings.NewReader(uploadData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Host", "picupload.weibo.com")
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	request.Header.Add("Referer", "https://weibo.com/u/"+uniqueid+"/home?wvr=5")
	request.Header.Add("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(request)
	if err != nil {
		logs.Error(err)
		return "", err
	}
	defer resp.Body.Close()

	queryParams := resp.Request.URL.RawQuery
	index := strings.Index(queryParams, `pid=`)
	pid := queryParams[index+4 : len(queryParams)]
	return pid, nil
}

func SendWeiBo(client *http.Client, uniqueid, pid, text string) error {
	//发微博
	sendURL := "https://weibo.com/aj/mblog/add?ajwvr=6&__rnd=" + fmt.Sprintf("%v", time.Now().UnixNano()/1e6)
	sendData := url.Values{}
	sendData.Add("location", "v6_content_home")
	sendData.Add("text", text)
	sendData.Add("appkey", "")
	sendData.Add("style_type", "1")
	sendData.Add("pic_id", pid)
	sendData.Add("tid", "")
	sendData.Add("pdetail", "")
	sendData.Add("mid", "")
	sendData.Add("isReEdit", "false")
	sendData.Add("rank", "0")
	sendData.Add("rankid", "")
	sendData.Add("module:", "stissue")
	sendData.Add("pub_source", "main_")
	sendData.Add("pub_type", "dialog")
	sendData.Add("isPri", "0")
	sendData.Add("_t", "0")

	request, err := http.NewRequest("POST", sendURL, strings.NewReader(sendData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	request.Header.Add("Referer", "https://weibo.com/u/"+uniqueid+"/home")
	_, err = client.Do(request)
	if err != nil {
		logs.Error(err)
		return err
	}
	return nil

}
func encryptPassword(pubkey string, servertime string, nonce string, password string) string {
	pub := rsa.PublicKey{
		N: string2big(pubkey),
		E: 65537, // 10001是十六进制数，65537是它的十进制表示
	}

	// servertime、nonce之间加\t，然后在\n ,和password拼接
	encryString := servertime + "\t" + nonce + "\n" + password

	// 拼接字符串加密
	encryResult, _ := rsa.EncryptPKCS1v15(rand.Reader, &pub, []byte(encryString))
	return hex.EncodeToString(encryResult)
}

func string2big(s string) *big.Int {
	ret := new(big.Int)
	ret.SetString(s, 16) // 将字符串转换成16进制
	return ret
}
