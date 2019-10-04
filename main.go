package main

import (
	"os"
	"strconv"
	"time"

	"./common"
	"./models"
	"./weibo"
	"github.com/astaxie/beego/logs"
)

func main() {
	Clock()
}

func Clock() {
	Spider()
	DownLoadPic()
	SendWB()

	for {
		now := time.Now()
		// 计算下一个零点
		next := now.Add(time.Hour * 24)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		t := time.NewTimer(next.Sub(now))
		<-t.C
		Spider()
	}
}

func GetPath() (string, error) {
	defaultDir, err := os.Getwd()
	if err != nil {
		logs.Error(err)
		return "", err
	}

	return defaultDir + "/" + models.GetToday(), nil
}

func Spider() error {
	list, err := common.GetRankList()
	if err != nil {
		logs.Error(err)
		return err
	}
	for k, v := range *list {
		exist, err := models.IsExist(v.ID)
		if err != nil {
			logs.Error(err)
			return err
		}
		if exist == true {
			continue
		}
		origin, regular, err := common.GetPicURL(v.ID)
		if err != nil {
			logs.Error(err)
			return err
		}
		for key, value := range *origin {

			err := models.InsertPicInfo(&v, value, (*regular)[key], k, key)
			if err != nil {
				logs.Error(err)
				return err
			}
		}

	}
	return nil
}

func DownLoadPic() error {
	localPath, err := GetPath()
	if err != nil {
		logs.Error(err)
		return err
	}
	err = os.Mkdir(localPath, os.ModePerm)
	if err != nil {
		logs.Error(err)
		return err
	}

	idList, err := models.GetTodayPic()
	if err != nil {
		logs.Error(err)
		return err
	}
	for _, v := range *idList {
		urlList, err := models.SelectPic(v)
		if err != nil {
			logs.Error(err)
			return err
		}
		for key, value := range *urlList {

			Path := localPath + "/" + v + "_" + strconv.Itoa(key) + ".jpg"
			err := common.DownLoadPic(v, value, Path)
			if err != nil {
				logs.Error(err)
				return err
			}

		}
	}
	return nil
}

func SendWB() error {
	client, uniqueid, err := weibo.SignIn()
	if err != nil {
		logs.Error(err)
		return err
	}
	idList, err := models.GetTodayPic()
	if err != nil {
		logs.Error(err)
		return err
	}
	localPath, err := GetPath()
	if err != nil {
		logs.Error(err)
		return err
	}
	for _, v := range *idList {
		picinfo, err := models.GetPicDetail(v)
		if err != nil {
			logs.Error(err)
			return err
		}
		for i := 0; i < (picinfo.Num/9)+1; i++ {
			count := 9
			if i == picinfo.Num/9 {
				count = picinfo.Num - (9 * (picinfo.Num / 9))
			}
			var pids string
			for k := 0; k < count; k++ {
				path := localPath + "/" + v + "_" + strconv.Itoa(9*i+k) + ".jpg"
				pid, err := weibo.UploadPic(client, uniqueid, path)
				if err != nil {
					logs.Error(err)
					return err
				}
				pids = pids + "|" + pid
			}
			pids = pids[1:len(pids)]
			text := models.GetToday() + `
			第` + strconv.Itoa(picinfo.Rank) + `
			标题：` + picinfo.Title + `
			共` + strconv.Itoa(picinfo.Num) + `张
			第(` + strconv.Itoa(i+1) + `/` + strconv.Itoa((picinfo.Num/9)+1) + `)批`
			err = weibo.SendWeiBo(client, uniqueid, pids, text)
			if err != nil {
				logs.Error(err)
				return err
			}
		}

	}

	return nil
}
