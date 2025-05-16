package main

import (
   "os"
   //"io"
   "fmt"
   "time"
   //"bytes"
   "strings"
   // "strconv"
   "net/http"
   "encoding/json"
   // "path/filepath"
   // "mime/multipart"
   // "github.com/joho/godotenv"
)

// Details represents the details of a model
type Details struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// Model represents a single model's information 模型資訊結構
type Model struct {
   Name        string    `json:"name"`
	Model			string    `json:"model"`
   ModifiedAt  time.Time `json:"modified_at"`
   Size        int64     `json:"size"`
   Digest      string    `json:"digest"`
	Details		Details   `json:"details"`
   Description string    `json:"description,omitempty"`
	SizeGB		string	`json:"size_gb,omitempty"`	// 轉換後的大小
	ModiTime	string		`json:"modi_time,omitempty"`	// 轉換後的時間
}

// ModelsWrapper wraps the models array
type ModelsWrapper struct {
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

type OllamaClient struct {
   URL		string		`json:"url"`		// Ollama 服務的 URL
   Models	[]Model		`json:"models"`   // 模型列表回應結構
}

// 獲取所有可用模型
func(app *OllamaClient) getModels() ([]Model, error) {
   resp, err := http.Get(app.URL + "/api/tags")
   if err != nil {
      return nil, fmt.Errorf("連接到 Ollama 服務失敗: %v", err)
   }
   defer resp.Body.Close()
   if resp.StatusCode != http.StatusOK {
      return nil, fmt.Errorf("獲取模型列表失敗，狀態碼: %d", resp.StatusCode)
   }
   var listResp ModelsWrapper
   if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
      return nil, fmt.Errorf("解析回應失敗: %v", err)
   }
   return listResp.Models, nil
}

// 顯示所有可用模型 <option value="gpt-4">GPT-4</option>
func(app *OllamaClient) ListModelsFromWeb(w http.ResponseWriter, r *http.Request) {
   s := []string{}
   for _, model := range app.Models {
      val := strings.ReplaceAll(strings.ToLower(model.Name), " ", "-")
      s = append(s, fmt.Sprintf("<option value=\"%s\">%s</option>", val, model.Name))
	}
	jsonData, err := json.Marshal(s)
	if err != nil {
		fmt.Println("序列化 JSON 失敗:", err)
		return	
	}
	// 回傳用戶消息，實際處理會通過SSE進行
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)	
   w.Write(jsonData)	// 將 JSON 數據寫入響應
}


func(app *OllamaClient) AddRouter(router *http.ServeMux) {
   router.HandleFunc("GET /models", app.ListModelsFromWeb)                // 列出工具列表
   // router.HandleFunc("/update/overwrite/{type}", app.OverWriteDataFromWeb)      // 更新檔案內容
}

func NewOllamaClient()(*OllamaClient) {
   app := &OllamaClient{
      URL: os.Getenv("OllamaUrl"),
   }
   if app.URL == "" {
      fmt.Printf("未設定 envfile 中的 OllamaUrl 參數")
      return nil
   }
	var err error
   app.Models, err = app.getModels()
   if err != nil {
      fmt.Printf("獲取模型列表失敗: %s", err.Error())
      return nil
   }

   for i, model := range app.Models {
      size := float64(model.Size) / (1024 * 1024 * 1024)
      app.Models[i].SizeGB = fmt.Sprintf("%.2f GB", size)
      app.Models[i].ModiTime = model.ModifiedAt.Format("2006-01-02 15:04:05")
   }
   return app
}

/*
func main() {
   currentDir, err := os.Getwd()
   if err != nil {
      fmt.Println(err.Error())
      return
   }
   if err := godotenv.Load(currentDir + "/envfile"); err != nil {
      fmt.Println(err.Error())
      return
   }
   if olm := NewOllamaClient(); olm != nil {
		olm.ListModelsFromWeb()
	}
}

// 產生回應（非串流模式）
func generateResponse(modelName, prompt string, files []string) (string, error) {
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
*/