package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/openai/openai-go"
)


const INPUT_FILE = "topics.txt"
const OUTPUT_DIR = "output"
const MAX_WORKERS = 3
const PROMPT = `
You are given a topic keyword and you are able to summarize the topic in a few sentences.
You should include a general summary of the topic.
You should include a list of the most important points of the topic.
You should include a list of the most important questions about the topic.
You should include a list of the most important answers to the questions.
You should include a list of the most important sources of information about the topic.
You should include a list of the most important links to the sources of information.

`


type ResearchData struct {
	Content string
	Date time.Time
	Took time.Duration
	Topic string	
	WorkerId int
}

var client = openai.NewClient()


func errCheck(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		panic(err)
	}	
}

func writeFileWorker(outputChan chan ResearchData) {
	err := os.MkdirAll(OUTPUT_DIR, 0755)
	errCheck(err)
	for researchData := range outputChan {
		outFile := strings.ReplaceAll(researchData.Topic, " ", "_") + ".json"
		fmt.Println("Writing to file: ", outFile)
		f, err := os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		errCheck(err)
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", "  ") // Pretty print
		err = encoder.Encode(researchData)
		errCheck(err)
		f.Close()
	}
}
	

func researchWorker(researchChan chan ResearchData, topicsChan chan string, fileMutex *sync.Mutex, processedTopics *[]string, workerId int) {	
	defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Worker %d panicked: %v\n", workerId, r)
        }
    }()

	for researchTopic := range topicsChan {
		start := time.Now()
		fmt.Println("Researching topic: ", researchTopic)
		chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(PROMPT + researchTopic),
			},
			Model: openai.ChatModelGPT4o,
		})

		errCheck(err)
		researchChan <- ResearchData{
			Content: chatCompletion.Choices[0].Message.Content,
			Date: time.Now(),
			Took: time.Since(start),
			Topic: researchTopic,
			WorkerId: workerId,
		}
		fileMutex.Lock()
		fmt.Println("Adding to list of processed topics: ", researchTopic)

		*processedTopics = append(*processedTopics, researchTopic)
		fileMutex.Unlock()
	}
}	

func readFile(processedTopics *[]string, topicsChan chan string, filename string) {
	fmt.Println("Reading file: ", filename)
	content, err := os.ReadFile(filename)
	errCheck(err)
	topics := strings.Split(string(content), "\n")

	for i := range topics {
		topic := strings.TrimSpace(topics[i])	
		if topic == "" {
			continue
		}
		for _, processedTopic := range *processedTopics {
			if topic == processedTopic {
				fmt.Println("Skipping previously processed topic: ", topic)
				continue
			}
		}
		topicsChan <- topic
	}
	fmt.Println("Finished reading file: ", filename)

}


func checkFileWorker(topicsChan chan string, processedTopics *[]string, filename string) {
	fmt.Println("Checking file: ", filename)
	checkFileTicker := time.NewTicker(5000 * time.Millisecond)
	for {
		select {
		case <-checkFileTicker.C:
			readFile(processedTopics, topicsChan, filename)
		}
	}
}

func main() {
	researchChan := make(chan ResearchData)
	topicsChan := make(chan string)

	processedTopics := []string{}

	go checkFileWorker(topicsChan, &processedTopics, INPUT_FILE)
	go writeFileWorker(researchChan)
	fileMutex := sync.Mutex{}
	for i := 0; i < MAX_WORKERS; i++ {
		go researchWorker(researchChan, topicsChan, &fileMutex, &processedTopics, i)
	}	
	select {}
}