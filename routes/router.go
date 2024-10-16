package routes

import (
	"github.com/gin-gonic/gin"

	"qa/api"
	v1 "qa/api/v1"
	"qa/middleware/auth"
)

// NewRouter 路由配置
func NewRouter() *gin.Engine {
	r := gin.Default()

	// 主页
	r.GET("/", api.Index)

	v1Group := r.Group("/api/v1")
	{
		// 注册
		v1Group.POST("/user/register", v1.UserRegister)
		// 登录
		v1Group.POST("/user/login", v1.UserLogin)

		// 获取首页推荐列表
		v1Group.GET("/questions", v1.FindQuestions)
		// 获取问题热榜
		v1Group.GET("/hot_questions", v1.FindHotQuestions)
		// 获取回答列表
		v1Group.GET("/questions/:qid/answers", v1.FindAnswers)

		// 可选token
		jwtSelect := v1Group.Group("")
		jwtSelect.Use(auth.JwtWithAnonymous())
		{
			// 查看单个问题
			jwtSelect.GET("/questions/:qid", v1.FindOneQuestion)
			// 查看单个回答
			jwtSelect.GET("/questions/:qid/answers/:aid", v1.FindAnswer)
		}

		// 需要登录权限
		jwt := v1Group.Group("")
		jwt.Use(auth.JwtRequired())
		{
			// 查看个人信息
			jwt.GET("/user/me", v1.UserMe)
			// 退出登录
			jwt.POST("/user/logout", v1.Logout)
			// 查看个人发布问题
			jwt.GET("/user/questions", v1.GetUserQuestions)
			// 查看个人发布回答
			jwt.GET("/user/answers", v1.GetUserAnswers)
			// 查看点赞回答列表
			jwt.GET("/user/awesomes", v1.Awesomes)

			// 发布问题
			jwt.POST("/questions", v1.QuestionAdd) //每次新发布问题，就在函数中自动触发 回答问题，这个回答不是人输入的，是调用llm进行生成的，并放在这个问题的回答中，
			// 修改问题
			jwt.PUT("/questions/:qid", v1.EditQuestion) //这里在每个下面中也有：每次进行问题的修改后，都进行RAG技术的新的回答生成，
			// 删除问题
			jwt.DELETE("/questions/:qid", v1.DeleteQuestion)

			// 回答问题   需要标注AI生成，以区分每个问题的回答是由AI生成的还是由人工又删除进行添加或者修改的回答
			jwt.POST("/questions/:qid/answers", v1.AddAnswer) // 这个下两层的函数的使用放在了 发布问题 函数的内部，//这个可以人直接用，LLM间接用
			// 修改回答
			jwt.PUT("/questions/:qid/answers/:aid", v1.UpdateAnswer) // 每个回答一个id 和问题id没有直接关系，这里可以手动修改，//这个必须人用
			// 删除回答
			//jwt.DELETE("/questions/:qid/answers/:aid", v1.DeleteAnswer) // 每次删除，这个问题就没有回答了，可以空着，也可以人工进行回答问题，
			// 点赞回答
			jwt.POST("/answers/:aid/voters", v1.Voter)
		}
	}

	return r
}
