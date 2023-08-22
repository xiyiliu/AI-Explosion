package controller

import (
	"TIKTOK/dao"
	"TIKTOK/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

// 评论列表信息获取
type CommentListResponse struct {
	StatusCode  int64                 `json:"status_code,omitempty"`
	StatusMsg   string                `json:"status_msg,omitempty"`
	CommentList []service.CommentInfo `json:"comment_list,omitempty"`
}

// 发表评论返回参数
type CommentActionResponse struct {
	StatusCode int64               `json:"statusCode,omitempty"`
	StatusMsg  string              `json:"statusMSG,omitempty"`
	Comment    service.CommentInfo `json:"comment,omitempty"`
}

func CommentAction(c *gin.Context) {
	id, _ := c.Get("userId")
	userid, _ := id.(string)
	userId, err := strconv.ParseInt(userid, 10, 64)
	log.Printf("err:%v", err)
	log.Printf("userId:%v", userId)
	
	if err != nil {
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "comment userId json invalid",
		})
		log.Println("CommentController-Comment_Action: return comment userId json invalid") //函数返回userId无效
		return
	}
	
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	
	if err != nil {
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "comment videoId json invalid",
		})
		log.Println("CommentController-Comment_Action: return comment videoId json invalid") 
		return
	}
	
	actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 32)
	
	if err != nil || actionType < 1 || actionType > 2 {
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "comment actionType json invalid",
		})
		log.Println("CommentController-Comment_Action: return actionType json invalid") 
		return
	}
	//调用service层评论函数
	commentService := new(service.CommentService) 
	if actionType == 1 {                          
		content := c.Query("comment_text")

		
		var sendComment dao.Comment
		sendComment.UserId = userId
		sendComment.VideoId = videoId
		sendComment.Content = content
		timeNow := time.Now()
		sendComment.CreateTime = timeNow
		
		commentInfo, err := commentService.Send(sendComment)
		
		if err != nil {
			c.JSON(http.StatusOK, CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "send comment failed",
			})
			log.Println("CommentController-Comment_Action: return send comment failed") //发表失败
			return
		}

		
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: 0,
			StatusMsg:  "send comment success",
			Comment:    commentInfo,
		})
		log.Println("CommentController-Comment_Action: return Send success") 
		return
	} else { 
		commentId, err := strconv.ParseInt(c.Query("comment_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "delete commentId invalid",
			})
			log.Println("CommentController-Comment_Action: return commentId invalid") 
			return
		}
		
		err = service.CommentService.DeleteComment()
		if err != nil { 
			str := err.Error()
			c.JSON(http.StatusOK, CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  str,
			})
			log.Println("CommentController-Comment_Action: return delete comment failed") 
			return
		}
		
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: 0,
			StatusMsg:  "delete comment success",
		})

		log.Println("CommentController-Comment_Action: return delete success") 
		return
	}
}

type Response struct {
	StatusCode int
	StatusMsg  string
}


func CommentList(c *gin.Context) {
	
	
	id, _ := c.Get("userId")
	userid, _ := id.(string)
	userId, err := strconv.ParseInt(userid, 10, 64)
	
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: -1,
			StatusMsg:  "comment videoId json invalid",
		})
		log.Println("CommentController-Comment_List: return videoId json invalid") //视频id格式有误
		return
	}
	log.Printf("videoId:%v", videoId)

	
	commentService := new(service.CommentServiceImpl)
	commentList, err := commentService.GetList(videoId, userId)
	
	if err != nil { 
		c.JSON(http.StatusOK, CommentListResponse{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		})
		log.Println("CommentController-Comment_List: return list false") 
		return
	}

	c.JSON(http.StatusOK, CommentListResponse{
		StatusCode:  0,
		StatusMsg:   "get comment list success",
		CommentList: commentList,
	})
	log.Println("CommentController-Comment_List: return success") 
	return
}
