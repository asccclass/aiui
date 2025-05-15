package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ollamaURL = "http://10.109.190.12"
)

// 模型資訊結構
type Model struct {
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ModifiedAt  time.Time `json:"modified_at"`
	Digest      string    `json:"digest"`
	Description string    `json:"description,omitempty"`
}

// 模型列表回應結構
type ListModelsResponse struct {
	Models []Model `json:"models"`
}

// 生成請求結構
type GenerateRequest struct {
	Model    string   `json:"model"`
	Prompt   string   `json:"prompt"`
	System   string   `json:"system,omitempty"`
	Format   string   `json:"format,omitempty"`
	Stream   bool     `json:"stream,omitempty"`
	Options  Options  `json:"options,omitempty"`
	Images   []string `json:"images,omitempty"`
	Context  []int    `json:"context,omitempty"`
	Template string   `json:"template,omitempty"`
}

// 生成選項結構
type Options struct {
	Temperature      float64 `json:"temperature,omitempty"`
	TopP             float64 `json:"top_p,omitempty"`
	TopK             int     `json:"top_k,omitempty"`
	NumPredict       int     `json:"num_predict,omitempty"`
	NumKeep          int     `json:"num_keep,omitempty"`
	Seed             int     `json:"seed,omitempty"`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64 `json:"presence_penalty,omitempty"`
	Mirostat         int     `json:"mirostat,omitempty"`
	MirostatEta      float64 `json:"mirostat_eta,omitempty"`
	MirostatTau      float64 `json:"mirostat_tau,omitempty"`
	Stop             string  `json:"stop,omitempty"`
}

// 生成回應結構
type GenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// 獲取所有可用模型
func getModels() ([]Model, error) {
	resp, err := http.Get(ollamaURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("連接到 Ollama 服務失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("獲取模型列表失敗，狀態碼: %d", resp.StatusCode)
	}

	var listResp ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("解析回應失敗: %v", err)
	}

	return listResp.Models, nil
}

// 顯示所有可用模型
func displayModels(models []Model) {
	fmt.Println("=== 可用模型列表 ===")
	for i, model := range models {
		size := float64(model.Size) / (1024 * 1024 * 1024)
		fmt.Printf("%d. %s (%.2f GB) - 修改時間: %s\n",
			i+1, model.Name, size, model.ModifiedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("==================")
}

// 產生回應（非串流模式）
func generateResponse(modelName, prompt string, files []string) (string, error) {
	// 準備請求
	reqBody := GenerateRequest{
		Model:  modelName,
		Prompt: prompt,
		Stream: false,
	}

	// 如果有上傳文件，將文件內容添加到提示
	if len(files) > 0 {
		var fileContents []string
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				return "", fmt.Errorf("讀取文件 %s 失敗: %v", file, err)
			}
			fileContents = append(fileContents, fmt.Sprintf("\n文件 '%s' 內容:\n%s", filepath.Base(file), string(content)))
		}
		
		// 將文件內容附加到提示
		if len(fileContents) > 0 {
			reqBody.Prompt += "\n\n以下是相關文件內容，請參考這些內容回答我的問題:" + strings.Join(fileContents, "\n\n")
		}
	}

	// 將請求轉為 JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化請求失敗: %v", err)
	}

	// 發送請求
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("發送請求失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("生成回應失敗，狀態碼: %d", resp.StatusCode)
	}

	// 解析回應
	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("解析回應失敗: %v", err)
	}

	return genResp.Response, nil
}

// 讀取用戶輸入
func getUserInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

// 上傳檔案處理
func handleFileUpload() []string {
	files := []string{}
	
	for {
		fmt.Println("\n請選擇操作:")
		fmt.Println("1. 上傳檔案")
		fmt.Println("2. 完成上傳並繼續")
		
		choice := getUserInput("請輸入選項 (1-2): ")
		
		switch choice {
		case "1":
			path := getUserInput("請輸入檔案路徑: ")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("錯誤: 檔案 '%s' 不存在\n", path)
				continue
			}
			files = append(files, path)
			fmt.Printf("已添加檔案: %s\n", path)
		case "2":
			return files
		default:
			fmt.Println("無效選項，請重新選擇")
		}
	}
}

func main() {
	fmt.Println("=== Ollama 客戶端 ===")
	fmt.Printf("連接到 Ollama 服務: %s\n", ollamaURL)
	
	// 獲取可用模型
	models, err := getModels()
	if err != nil {
		log.Fatalf("獲取模型列表失敗: %v", err)
	}
	
	if len(models) == 0 {
		log.Fatalf("沒有找到可用模型")
	}
	
	// 顯示可用模型
	displayModels(models)
	
	// 選擇模型
	var modelIndex int
	for {
		input := getUserInput("請選擇模型序號 (1-" + strconv.Itoa(len(models)) + "): ")
		index, err := strconv.Atoi(input)
		if err != nil || index < 1 || index > len(models) {
			fmt.Println("無效選項，請重新選擇")
			continue
		}
		modelIndex = index - 1
		break
	}
	
	selectedModel := models[modelIndex].Name
	fmt.Printf("已選擇模型: %s\n", selectedModel)
	
	// 主要對話循環
	for {
		fmt.Println("\n=== 命令選單 ===")
		fmt.Println("1. 發送訊息")
		fmt.Println("2. 上傳檔案並發送訊息")
		fmt.Println("3. 更換模型")
		fmt.Println("4. 退出程式")
		
		choice := getUserInput("請選擇操作 (1-4): ")
		
		switch choice {
		case "1":
			prompt := getUserInput("請輸入您的訊息: ")
			if prompt == "" {
				fmt.Println("訊息不能為空")
				continue
			}
			
			fmt.Println("正在處理，請稍候...")
			response, err := generateResponse(selectedModel, prompt, nil)
			if err != nil {
				fmt.Printf("錯誤: %v\n", err)
				continue
			}
			
			fmt.Println("\n=== 模型回應 ===")
			fmt.Println(response)
			fmt.Println("=================")
			
		case "2":
			fmt.Println("=== 文件上傳 ===")
			files := handleFileUpload()
			
			if len(files) == 0 {
				fmt.Println("未上傳任何檔案")
				continue
			}
			
			fmt.Printf("已上傳 %d 個檔案\n", len(files))
			prompt := getUserInput("請輸入您的訊息: ")
			if prompt == "" {
				fmt.Println("訊息不能為空")
				continue
			}
			
			fmt.Println("正在處理，請稍候...")
			response, err := generateResponse(selectedModel, prompt, files)
			if err != nil {
				fmt.Printf("錯誤: %v\n", err)
				continue
			}
			
			fmt.Println("\n=== 模型回應 ===")
			fmt.Println(response)
			fmt.Println("=================")
			
		case "3":
			displayModels(models)
			for {
				input := getUserInput("請選擇模型序號 (1-" + strconv.Itoa(len(models)) + "): ")
				index, err := strconv.Atoi(input)
				if err != nil || index < 1 || index > len(models) {
					fmt.Println("無效選項，請重新選擇")
					continue
				}
				modelIndex = index - 1
				break
			}
			selectedModel = models[modelIndex].Name
			fmt.Printf("已切換至模型: %s\n", selectedModel)
			
		case "4":
			fmt.Println("程式已結束")
			return
			
		default:
			fmt.Println("無效選項，請重新選擇")
		}
	}
}
