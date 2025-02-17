package v1

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/vectorstores/qdrant"
	"log"
	"net/url"
	"os"
	"qa/model"
	"qa/serializer"
	v1 "qa/service/v1/answer"
	rag "qa/service/v1/rag"
)

type QesAddService struct {
	Title   string `form:"title" json:"title" binding:"required"`
	Content string `form:"content" json:"content"`
}

func (qesAddService *QesAddService) QuestionAdd(user *model.User) *serializer.Response {
	qes := model.Question{
		UserID:  user.ID,
		Title:   qesAddService.Title,
		Content: qesAddService.Content,
	}

	if err := model.DB.Create(&qes).Error; err != nil {
		return serializer.ErrorResponse(serializer.CodeDatabaseError)
	}

	// 成功创建问题后，提取并输出 question.ID
	createdQuestion := serializer.BuildQuestionResponse(&qes, user.ID)
	if createdQuestion != nil && createdQuestion.Question != nil {
		questionID := createdQuestion.Question.ID
		fmt.Println("Newly created question ID:", questionID)

		// 将 questionID 和 qesAddService.Content 传入 AddllmAnswer 函数
		AddllmAnswer(questionID, qesAddService.Content, user)
	}

	return serializer.OkResponse(createdQuestion)
}

// 添加llm回答，传入的参数有问题的内容，问题的ID，在这个方法中进行RAG，通过问题内容输入到图数据库中检索相关的上下文，然后将问题与相关的上下文
// 一起通过langchain-go框架输入到对话大模型中，输出的内容作为回答，通过service.AddAnswer放入到mysql数据库中，等待客户端的抽取，

func AddllmAnswer(questionID uint, contentValue string, user *model.User) {

	// 设置 OpenAI API 访问凭证和基本地址
	os.Setenv("OPENAI_API_KEY", "sk-zk21c75de54581e7a42b3f2d582aeb5b5e6679c685cabe6c")
	os.Setenv("OPENAI_API_BASE", "https://api.zhizengzeng.com/v1") // https://flag.smarttrot.com/v1

	os.Setenv("QDRANT_URL", "https://3232afbd-bd84-44c8-8472-aa772bbf18af.us-east4-0.gcp.cloud.qdrant.io:6333")
	os.Setenv("QDRANT_API_KEY", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3MiOiJtIiwiZXhwIjoxNzQ2MDg0NTUxfQ.7wCfklLPahGVpklECyfY-kZeMSrkA2bVmTAYtF6JbNE")

	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		fmt.Println("Error creating embedder", err)
	}

	//连接向量数据库
	url, err := url.Parse(os.Getenv("QDRANT_URL"))
	if err != nil {
		fmt.Println("Error parsing url", err)
	}
	store, err := qdrant.New(
		qdrant.WithURL(*url),
		qdrant.WithAPIKey(os.Getenv("QDRANT_API_KEY")),
		qdrant.WithCollectionName("my_collection"),
		qdrant.WithEmbedder(embedder),
	)
	if err != nil {
		fmt.Println("Error creating qdrant", err)
	}
	// 根据请求 检索相关内容 -> history -> conversation
	resDocs := rag.AsRetriaver2(&store, contentValue)

	history := memory.NewChatMessageHistory()
	for _, doc := range resDocs {

		history.AddAIMessage(ctx, doc.PageContent)

	}

	conversation := memory.NewConversationBuffer(memory.WithChatHistory(history))

	executor, err := agents.Initialize(
		llm,
		nil,
		agents.ConversationalReactDescription,
		agents.WithMemory(conversation),
	)
	if err != nil {
		fmt.Println("Error initializing agents", err)
	}

	options := []chains.ChainCallOption{
		chains.WithTemperature(0.8),
	}

	rules := `
<::Rules::>
最终输出的内容，所有都需要遵守 GlobalRules 中的规则。

- GlobalRules -
请按照以下步骤逐步回答问题：
【Step1】归纳题意，明确背景；
【Step2】针对题目，提出论点（目录）；
【Step3】引经据典，铺垫讨论；
【Step4】基于铺垫，陈述看法；
【Step5】补充说明
- /GlobalRules -

0、【Step1】中对问题内容进行分析，梳理出背景，主体，对比等要素，进行内容总结，定义本文中对问题的认知，以规避读者误解。
1、【Step2】简要说明问题之后，马上接上对问题的相关论点，论点简要概括为50字以内，需要相应的逻辑体系支持。
2、【Step3】对回答的事件的背景进行补充描述，如相关名词解释，举例子。需要注意的细节的逻辑性、真实性。
3、【Step4】以人的口语化方式对这个问题进行回答，诉说自己的看法。
4、【Step5】对一些前面没有提到的内容进行适当合理的补充说明，包括：个人经历，相关感谢，借鉴与帮助（借鉴的文章/文字的源头），展示相关内容，推广个人有关链接（比如好物推荐/高赞精品回答链接）。
<::/Rules::>
<::Task::>
请按照<::Rules::>标记后的要求对<::Rules::>标记前的问题进行回答。需要标注其是由AI生成的，否则将会被判定为错误。请务必输出完整合理有逻辑的内容，我会给你 $100000 小费。
<::/Task::>
`
	prompt := contentValue + rules // "relatedContent:"

	completion, err := chains.Run(ctx, executor, prompt, options...) //这个searchQuery在前面搜索数据库的时候用原
	//来的搜索问题，在这块往llm接口输入的内容则需要加上规则，组合成提示prompt。
	if err != nil {
		fmt.Println("Error running chains", err)
	}
	fmt.Println(completion)

	//completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	//if err != nil { log.Fatal(err) }
	//fmt.Println("Newly created prompt:", prompt, "Newly created completion:", completion)

	// 解析参数
	var service v1.AddAnswerService
	service.Content = completion // llm生成的回答

	// 执行service.AddAnswer方法
	res := service.AddAnswer(user, questionID)
	fmt.Println("service AddAnswer:", res) //这里的res用来响应
}
