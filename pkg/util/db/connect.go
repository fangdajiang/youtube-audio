package db

import (
	"database/sql"
	"fmt"
	"youtube-audio/pkg/util/log"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectLocal() {
	// 设置数据库连接参数
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/audio-script")
	if err != nil {
		log.Fatalf("open local db error:%v", err)
	}
	defer db.Close()

	// 执行查询
	rows, err := db.Query("SELECT * FROM account")
	if err != nil {
		log.Fatalf("query local db error:%v", err)
	}
	log.Debugf("rows:%v", rows)

	// 循环遍历结果集
	for rows.Next() {
		var id int
		var userId string
		err = rows.Scan(&id, &userId)
		if err != nil {
			log.Fatalf("scan rows error:%v", err)
		}
		fmt.Println(id, userId)
	}
	err = rows.Err()
	if err != nil {
		log.Fatalf("rows err:%v", err)
	}

}
