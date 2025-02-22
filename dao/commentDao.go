package dao

import (
	"TIKTOK/config"
	"errors"
	"fmt"
	"log"
	"time"
)

type Comment struct {
	Id         int64
	UserId     int64
	VideoId    int64
	Content    string
	CreateTime time.Time
	Cancel     int
}

func (Comment) TableName() string {
	return "comments"
}

func Count(VideoId int64) (int64, error) {
	var count int64
	err := Db.Model(Comment{}).Where(map[string]interface{}{"video_id": VideoId, "cancel": config.ValidComment}).Count(&count).Error
	if err != nil {
		return -1, errors.New("find comments count failed")
	}
	return count, nil
}

func CommentIdList(VideoId int64) ([]string, error) {
	var commentIdList []string
	err := Db.Model(Comment{}).Select("id").Where("video_id=?", VideoId).Find(&commentIdList).Error
	if err != nil {
		return []string{}, fmt.Errorf("fail to get commentidlist :%s", err.Error())
	}
	return commentIdList, nil
}

func InsertComment(comment Comment) (Comment, error) {

	err := Db.Model(Comment{}).Create(&comment).Error
	if err != nil {
		return Comment{}, err
	}
	log.Println("insert comment successfully")
	return comment, nil

}

func DeleteComment(Id int64) error {

	result := Db.Model(Comment{}).Where(map[string]interface{}{"id": Id, "cancel": config.ValidComment}).Find(&Comment{})
	if result.RowsAffected == 0 {
		return errors.New("comment doesn't exist")
	}

	err := Db.Model(Comment{}).Where("id=?", Id).Update("cancel", config.ValidComment).Error
	if err != nil {
		return errors.New("comment exists but failed to delete")
	}
	log.Println("delete comment successfully")
	return nil
}

func GetCommentList(VideoId int64) ([]Comment, error) {
	var commentList []Comment
	result := Db.Model(Comment{}).Where(map[string]interface{}{"video_id": VideoId, "cancel": config.ValidComment}).Order("create_date desc").Find(&commentList)
	if result.RowsAffected == 0 {
		log.Println("no comment exists")
		return nil, nil
	}
	if result.Error != nil {
		log.Println("failed to get commentList")
		return nil, nil
	}
	return commentList, nil

}
