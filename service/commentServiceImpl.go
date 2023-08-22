package service

import (
	"TIKTOK/config"
	"TIKTOK/dao"
	rabbitmq "TIKTOK/middlewear/rabbitMQ"
	"TIKTOK/middlewear/redis"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"
)

type CommentServiceImpl struct {
	UserService
}

func (c CommentServiceImpl) CountFromVideoId(videoId int64) (int64, error) {

	cnt, err := redis.RdbVCid.SCard(redis.Ctx, strconv.FormatInt(videoId, 10)).Result()
	if err != nil {
		//return 0, err
		log.Println("count from redis error:", err)
	}
	log.Println("comment count redis :", cnt)

	if cnt != 0 {
		return cnt - 1, nil
	}

	cntDao, err1 := dao.Count(videoId)
	log.Println("comment count dao :", cntDao)
	if err1 != nil {
		log.Println("comment count dao err:", err1)
		return 0, nil
	}

	go func() {

		cList, _ := dao.CommentIdList(videoId)

		_, _err := redis.RdbVCid.SAdd(redis.Ctx, strconv.Itoa(int(videoId)), config.DefaultRedisValue).Result()
		if _err != nil {
			log.Println("redis save one vId - cId 0 failed")
			return
		}

		_, err := redis.RdbVCid.Expire(redis.Ctx, strconv.Itoa(int(videoId)),
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			log.Println("redis save one vId - cId expire failed")
		}

		for _, commentId := range cList {
			insertRedisVideoCommentId(strconv.Itoa(int(videoId)), commentId)
		}
		log.Println("count comment save ids in redis")
	}()

	return cntDao, nil
}

func (c CommentServiceImpl) Send(comment dao.Comment) (CommentInfo, error) {

	var commentInfo dao.Comment
	commentInfo.VideoId = comment.VideoId
	commentInfo.UserId = comment.UserId
	commentInfo.Content = comment.Content
	commentInfo.Cancel = config.ValidComment
	commentInfo.CreateTime = comment.CreateTime

	commentRtn, err := dao.InsertComment(commentInfo)
	if err != nil {
		return CommentInfo{}, err
	}

	impl := UserServiceImpl{
		FollowService: &FollowServiceImp{},
	}
	userData, err2 := impl.GetUserByIdWithCurId(comment.UserId, comment.UserId)
	if err2 != nil {
		return CommentInfo{}, err2
	}

	commentData := CommentInfo{
		Id:         commentRtn.Id,
		UserInfo:   userData,
		Content:    commentRtn.Content,
		CreateDate: commentRtn.CreateTime.Format(config.DateTime),
	}

	go func() {
		insertRedisVideoCommentId(strconv.Itoa(int(comment.VideoId)), strconv.Itoa(int(commentRtn.Id)))
		log.Println("send comment save in redis")
	}()

	return commentData, nil
}

func (c CommentServiceImpl) DeleteComment(commentId int64) error {

	n, err := redis.RdbCVid.Exists(redis.Ctx, strconv.FormatInt(commentId, 10)).Result()
	if err != nil {
		log.Println(err)
	}
	if n > 0 {
		vid, err1 := redis.RdbCVid.Get(redis.Ctx, strconv.FormatInt(commentId, 10)).Result()
		if err1 != nil { //没找到，返回err
			log.Println("redis find CV err:", err1)
		}
		//删除，两个redis都要删除
		del1, err2 := redis.RdbCVid.Del(redis.Ctx, strconv.FormatInt(commentId, 10)).Result()
		if err2 != nil {
			log.Println(err2)
		}
		del2, err3 := redis.RdbVCid.SRem(redis.Ctx, vid, strconv.FormatInt(commentId, 10)).Result()
		if err3 != nil {
			log.Println(err3)
		}
		log.Println("delete comment in Redis success:", del1, del2) //del1、del2代表删除了几条数据

		rabbitmq.RmqCommentDel.Publish(strconv.FormatInt(commentId, 10))
		return nil
	}
	//不在内存中，则直接走数据库删除
	return dao.DeleteComment(commentId)
}

// GetList
// 4、查看评论列表-返回评论list
func (c CommentServiceImpl) GetList(videoId int64, userId int64) ([]CommentInfo, error) {

	commentList, err := dao.GetCommentList(videoId)
	if err != nil {
		log.Println("CommentService-GetList: return err: " + err.Error()) //函数返回提示错误信息
		return nil, err
	}
	//当前有0条评论
	if commentList == nil {
		return nil, nil
	}

	//提前定义好切片长度
	commentInfoList := make([]CommentInfo, len(commentList))

	wg := &sync.WaitGroup{}
	wg.Add(len(commentList))
	idx := 0
	for _, comment := range commentList {
		//2.调用方法组装评论信息，再append
		var commentData CommentInfo
		//将评论信息进行组装，添加想要的信息,插入从数据库中查到的数据
		go func(comment dao.Comment) {
			oneComment(&commentData, &comment, userId)
			//3.组装list
			//commentInfoList = append(commentInfoList, commentData)
			commentInfoList[idx] = commentData
			idx = idx + 1
			wg.Done()
		}(comment)
	}
	wg.Wait()

	sort.Sort(CommentSlice(commentInfoList))

	go func() {

		cnt, err1 := redis.RdbVCid.SCard(redis.Ctx, strconv.FormatInt(videoId, 10)).Result()
		if err1 != nil { //若查询缓存出错，则打印log
			//return 0, err
			log.Println("count from redis error:", err)
		}

		if cnt > 0 {
			return
		}

		_, _err := redis.RdbVCid.SAdd(redis.Ctx, strconv.Itoa(int(videoId)), config.DefaultRedisValue).Result()
		if _err != nil {
			log.Println("redis save one vId - cId 0 failed")
			return
		}

		_, err2 := redis.RdbVCid.Expire(redis.Ctx, strconv.Itoa(int(videoId)),
			time.Duration(config.OneMonth)*time.Second).Result()
		if err2 != nil {
			log.Println("redis save one vId - cId expire failed")
		}
		//将评论id循环存入redis
		for _, _comment := range commentInfoList {
			insertRedisVideoCommentId(strconv.Itoa(int(videoId)), strconv.Itoa(int(_comment.Id)))
		}
		log.Println("comment list save ids in redis")
	}()

	log.Println("CommentService-GetList: return list success")
	return commentInfoList, nil
}

func insertRedisVideoCommentId(videoId string, commentId string) {
	//在redis-RdbVCid中存储video_id对应的comment_id
	_, err := redis.RdbVCid.SAdd(redis.Ctx, videoId, commentId).Result()
	if err != nil { //若存储redis失败-有err，则直接删除key
		log.Println("redis save send: vId - cId failed, key deleted")
		redis.RdbVCid.Del(redis.Ctx, videoId)
		return
	}
	//在redis-RdbCVid中存储comment_id对应的video_id
	_, err = redis.RdbCVid.Set(redis.Ctx, commentId, videoId, 0).Result()
	if err != nil {
		log.Println("redis save one cId - vId failed")
	}
}

// 此函数用于给一个评论赋值：评论信息+用户信息 填充
func oneComment(comment *CommentInfo, com *dao.Comment, userId int64) {
	var wg sync.WaitGroup
	wg.Add(1)
	//根据评论用户id和当前用户id，查询评论用户信息
	impl := UserServiceImpl{
		FollowService: &FollowServiceImp{},
	}
	var err error
	comment.Id = com.Id
	comment.Content = com.Content
	comment.CreateDate = com.CreateTime.Format(config.DateTime)
	comment.UserInfo, err = impl.GetUserByIdWithCurId(com.UserId, userId)
	if err != nil {
		log.Println("CommentService-GetList: GetUserByIdWithCurId return err: " + err.Error()) //函数返回提示错误信息
	}
	wg.Done()
	wg.Wait()
}

// CommentSlice 此变量以及以下三个函数都是做排序-准备工作
type CommentSlice []CommentInfo

func (a CommentSlice) Len() int { //重写Len()方法
	return len(a)
}
func (a CommentSlice) Swap(i, j int) { //重写Swap()方法
	a[i], a[j] = a[j], a[i]
}
func (a CommentSlice) Less(i, j int) bool { //重写Less()方法
	return a[i].Id > a[j].Id
}
