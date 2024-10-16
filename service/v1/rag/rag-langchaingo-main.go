package v1

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	aai "github.com/AssemblyAI/assemblyai-go-sdk"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/qdrant"
)

var searchQuery string

// cd /home/etuser/桌面/问答项目/rag-langchaingo/rag-langchaingo-main
// go run ./main.go "what is chatgpt?"
// export QDRANT_URL=http://localhost:6333 设置向量数据库相关的环境变量，
/*func main() {
	//get param in cli
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./main \"search query\"")
		return
	}

	searchQuery = os.Args[1]

	fmt.Println("Searching for:", searchQuery)

	chain := v1.Pipe(
		downloadVideo,
		transcribeToText,
		textToSplit,
		ExecuteChains,
	)

	_, err := chain("https://www.youtube.com/watch?v=BrsocJb-fAo")
	if err != nil {
		log.Fatal(err)
	}
}*/

func downloadVideo(url string) string {
	//validate if  file exists
	_, err := os.Stat(generateFileName(url))
	if err == nil {
		fmt.Println("File already exists")
		return generateFileName(url)
	}

	fileName := generateFileName(url)
	cmd := exec.Command("yt-dlp", "--verbose", "-o", fileName, "-f", "mp4", "--extract-audio", "--audio-format", "mp3", url)

	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text()) // Print each line of the output in real-time
		}
	}()

	cmd.Wait()

	return fileName
}

func generateFileName(url string) string {
	parts := strings.Split(url, "/")
	videoID := parts[len(parts)-1]

	re := regexp.MustCompile("[^a-zA-Z0-9]+")
	videoID = re.ReplaceAllString(videoID, "_")
	videoID = strings.ToLower(videoID)

	fileName := videoID + ".mp3"
	return fileName
}

func transcribeToText(audioURL string) string {
	fmt.Println("Transcribing audio to text...")
	fmt.Println("Audio URL: ", audioURL)
	filename := strings.ReplaceAll(audioURL, ".mp3", ".txt")
	fmt.Println("File name: ", filename)
	_, err := os.Stat(filename)
	if err == nil {
		fmt.Println("File already exists")
		return filename
	}

	client := aai.NewClient(os.Getenv("Transcribe_API_KEY"))

	eaderMp3, err := os.Open(audioURL)

	if err != nil {
		fmt.Println("Error opening file: ", err)
	}

	transcript, err := client.Transcripts.TranscribeFromReader(context.Background(), eaderMp3, nil)
	if err != nil {
		fmt.Println("Something bad happened:", err)
		os.Exit(1)
	}

	//save file txt
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file: ", err)
		return ""
	}

	defer f.Close()

	_, err = f.WriteString(*transcript.Text)
	if err != nil {
		fmt.Println("Error writing to file: ", err)
		return ""
	}

	fmt.Println("Transcription complete!")
	return filename
}

func textToSplit(transcribeFile string) []schema.Document {

	f, err := os.Open(transcribeFile)
	if err != nil {
		fmt.Println("Error opening file: ", err)
	}

	p := documentloaders.NewText(f)

	split := textsplitter.NewRecursiveCharacter()
	split.ChunkSize = 500
	split.ChunkOverlap = 20
	docs, err := p.LoadAndSplit(context.Background(), split)

	if err != nil {
		fmt.Println("Error loading document: ", err)
	}

	log.Println("Document loaded: ", len(docs))

	return docs
}

func asStorage(docs []schema.Document, embedder *embeddings.EmbedderImpl) *qdrant.Store {
	url, err := url.Parse(os.Getenv("QDRANT_URL"))
	if err != nil {
		fmt.Println("Error parsing url", err)
		return nil
	}

	store, err := qdrant.New(
		qdrant.WithURL(*url),
		qdrant.WithAPIKey(os.Getenv("QDRANT_API_KEY")),
		qdrant.WithCollectionName("youtube_transcript"),
		qdrant.WithEmbedder(embedder),
	)
	if err != nil {
		fmt.Println("Error creating qdrant", err)
		return nil
	}

	if len(docs) > 0 {
		_, err = store.AddDocuments(context.Background(), docs) //这个方法可以多次使用，也就是可以多次往数据库中存数据文件。因为
		if err != nil {                                         // store是个实际存在的库，不是一个临时变量，
			fmt.Println("Error adding documents", err)
			return nil
		}
	}

	return &store
}

func AsRetriaver2(store *qdrant.Store, contentValue string) []schema.Document {
	optionsVector := []vectorstores.Option{
		vectorstores.WithScoreThreshold(0.80),
	}

	retriever := vectorstores.ToRetriever(store, 10, optionsVector...)
	// search
	resDocs, err := retriever.GetRelevantDocuments(context.Background(), contentValue)

	if err != nil {
		fmt.Println("Error getting relevant documents", err)
		return nil
	}

	return resDocs
}

func AsRetriaver(store *qdrant.Store) []schema.Document {
	optionsVector := []vectorstores.Option{
		vectorstores.WithScoreThreshold(0.80),
	}

	retriever := vectorstores.ToRetriever(store, 10, optionsVector...)
	// search
	resDocs, err := retriever.GetRelevantDocuments(context.Background(), searchQuery)

	if err != nil {
		fmt.Println("Error getting relevant documents", err)
		return nil
	}

	return resDocs
}

func ExecuteChains(docs []schema.Document) error {

	llm, err := ollama.New(ollama.WithModel("llama2"))

	if err != nil {
		fmt.Println("Error creating llama2 model", err)
		return err
	}

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		fmt.Println("Error creating embedder", err)
		return err
	}

	store := asStorage(docs, embedder) //后面项目合并时，可以在每次回答问题时不用存文件，在AddllmAnswer()方法中只有
	//检索部分就可以，但是怎么给分开：需要想办法把*qdrant.Store给摘出来，

	resDocs := AsRetriaver(store)

	history := memory.NewChatMessageHistory()
	ctx := context.Background()
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
		return err
	}

	options := []chains.ChainCallOption{
		chains.WithTemperature(0.8),
	}
	res, err := chains.Run(ctx, executor, searchQuery, options...) //这个searchQuery在前面搜索数据库的时候用原
	//来的搜索问题，在这块往llm接口输入的内容则需要加上规则，组合成提示prompt。

	if err != nil {
		fmt.Println("Error running chains", err)
		return err
	}

	fmt.Println(res)

	return nil

}
