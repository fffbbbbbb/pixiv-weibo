package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/astaxie/beego/logs"
	_ "github.com/lib/pq"
)

type PicInfo struct {
	ID    string
	Title string
}

type PicDetail struct {
	Title string
	Rank  int
	Num   int
}

const (
	host     = "114.55.103.181"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "weibo"
)

var db *sql.DB

func init() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		logs.Error(err)
		return
	}
	if err = db.Ping(); err != nil {
		logs.Error(err)
		return
	}
}

func GetToday() string {
	now := time.Now()
	ftime := now.AddDate(0, 0, -1)
	return ftime.Format("20060102")

}

func IsExist(pid string) (bool, error) {
	sqlSta := "select count(*) from picinfo where pid=$1"
	var Num int
	err := db.QueryRow(sqlSta, pid).Scan(&Num)
	if err != nil {
		logs.Error("select count from table err", err)
		return true, err
	}
	if Num != 0 {
		return true, nil
	}
	return false, nil
}

func InsertPicInfo(pid *PicInfo, origin, regular string, rank, array int) error {

	sqlSta := "insert into picinfo values($1,$2,$3,$4,$5,$6,$7)"
	_, err := db.Exec(sqlSta, pid.ID, pid.Title, rank+1, array, origin, regular, GetToday())
	if err != nil {
		logs.Error(err)
		return err
	}
	return nil
}

func GetTodayPic() (idList *[]string, err error) {
	idList = new([]string)

	sqlSta := "select pid from picinfo where data=$1 and arr=0 order by rank desc"
	rows, err := db.Query(sqlSta, GetToday())
	if err != nil {
		logs.Error("Qurey err", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var pid string
		err = rows.Scan(&pid)
		if err != nil {
			logs.Error("rows.Next err", err)
			return
		}
		*idList = append(*idList, pid)
	}
	if err = rows.Err(); err != nil {
		logs.Error("rows err", err)
		return
	}
	return
}

func SelectPic(pid string) (urlList *[]string, err error) {
	urlList = new([]string)
	sqlSta := "select regular from picinfo where pid=$1 order by arr"
	rows, err := db.Query(sqlSta, pid)
	if err != nil {
		logs.Error("Qurey err", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var url string
		err = rows.Scan(&url)
		if err != nil {
			logs.Error("rows.Next err", err)
			return
		}
		*urlList = append(*urlList, url)
	}
	if err = rows.Err(); err != nil {
		logs.Error("rows err", err)
		return
	}
	return
}

func GetPicDetail(pid string) (*PicDetail, error) {
	var picinfo PicDetail
	sqlSta := "select distinct rank,title from picinfo where pid=$1"
	err := db.QueryRow(sqlSta, pid).Scan(&(picinfo.Rank), &(picinfo.Title))
	if err != nil {
		logs.Error("select count from table err", err)
		return nil, err
	}
	sqlSta = "select count(*) from picinfo where pid=$1"
	err = db.QueryRow(sqlSta, pid).Scan(&(picinfo.Num))
	if err != nil {
		logs.Error("select count from table err", err)
		return nil, err
	}
	return &picinfo, nil

}
