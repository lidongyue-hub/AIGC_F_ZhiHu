package v1

import (
	"qa/cache"
	"qa/model"
	"qa/serializer"
	"strconv"
)

// EditQuestionService 管理修改问题的服务
type EditQuestionService struct {
	Title   string `form:"title" json:"title" binding:"required"`
	Content string `form:"content" json:"content"`
}

// isQuestionOwner 判断用户是否拥有该问题
func isQuestionOwner(user *model.User, id uint) *serializer.Response {
	question, err := model.GetQuestion(id)
	if err != nil {
		return serializer.ErrorResponse(serializer.CodeQuestionIdError)
	}
	if question.UserID != user.ID {
		return serializer.ErrorResponse(serializer.CodeQuestionNotOwn)
	}
	return nil
}

// EditQuestion 修改问题
func (editQuestionService *EditQuestionService) EditQuestion(user *model.User, id uint) *serializer.Response {
	if err := isQuestionOwner(user, id); err != nil {
		return err
	}
	if question, err := model.UpdateQuestion(id, map[string]interface{}{
		"title":   editQuestionService.Title,
		"content": editQuestionService.Content,
	}); err != nil {
		return serializer.ErrorResponse(serializer.CodeDatabaseError)
	} else {
		if _, err := cache.RedisClient.HGet(cache.KeyHotQuestionTitle, strconv.Itoa(int(id))).Result(); err == nil {
			cache.RedisClient.HSet(cache.KeyHotQuestionTitle, strconv.Itoa(int(id)), question.Title)
		}
		AddllmAnswer(id, editQuestionService.Content, user) //每次进行问题的修改后，都进行RAG技术的新的回答生成，等于是给这个不变的问题ID增加了个新ID回答，

		return serializer.OkResponse(serializer.BuildQuestionResponse(question, user.ID))
	}

}

// DeleteQuestion 删除问题
func DeleteQuestion(user *model.User, id uint) *serializer.Response {
	if err := isQuestionOwner(user, id); err != nil {
		return err
	}
	err := model.DeleteQuestion(id)
	if err != nil {
		return serializer.ErrorResponse(serializer.CodeDatabaseError)
	}
	pipe := cache.RedisClient.TxPipeline()
	pipe.ZRem(cache.KeyHotQuestions, strconv.Itoa(int(id)))
	pipe.ZRem(cache.KeyHotQuestionTitle, strconv.Itoa(int(id)))
	pipe.HDel(cache.KeyHotAnswer, strconv.Itoa(int(id)))
	pipe.Exec()
	return serializer.OkResponse(nil)
}
