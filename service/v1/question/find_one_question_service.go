package v1

import (
	"qa/model"
	"qa/serializer"
)

func FindOneQuestion(id uint, uid uint) *serializer.Response {
	if question, err := model.GetQuestion(id); err == nil {
		return serializer.OkResponse(serializer.BuildQuestionResponse(question, uid))
	}
	return serializer.ErrorResponse(serializer.CodeQuestionIdError)
}
