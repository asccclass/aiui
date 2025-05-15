package main

import(
   "fmt"
   "time"
   "net/http"
   "encoding/json"
)

// 消息結構
type ChatMessage struct {
   Type    string `json:"type"`
   Content string `json:"content"`
}

// 模擬AI回應（實際應用中會調用真實的AI服務）
func simulateAIResponse(userMessage, model string) []ChatMessage {
   // 根據不同模型準備不同的回應
   var response string
   switch model {
   case "gpt-3.5":
	response = fmt.Sprintf("我是GPT-3.5，您說: \"%s\"。我的回應是：這是一個基於GPT-3.5的回應。", userMessage)
   case "gpt-4":
	response = fmt.Sprintf("我是GPT-4，您說: \"%s\"。我的回應是：這是一個基於GPT-4的高級回應。", userMessage)
   case "claude-3":
	response = fmt.Sprintf("我是Claude-3，您說: \"%s\"。我的回應是：這是一個基於Claude-3的溫暖對話。", userMessage)
   case "gemini-pro":
	response = fmt.Sprintf("我是Gemini Pro，您說: \"%s\"。我的回應是：這是一個基於Gemini Pro的創新回應。", userMessage)
   default:
	response = fmt.Sprintf("收到: \"%s\"。但我不確定使用哪個AI模型回應您。", userMessage)
   }

   // 將回應分割成小塊以模擬流式輸出
   chunks := []ChatMessage{}
   words := []rune(response)
   chunkSize := 5 // 每次發送5個字符

   for i := 0; i < len(words); i += chunkSize {
      end := i + chunkSize
      if end > len(words) {
         end = len(words)
      }
      chunk := string(words[i:end])
      chunks = append(chunks, ChatMessage{
         Type:    "chunk",
         Content: chunk,
      })
   }

   return chunks
}

func SSEChat(w http.ResponseWriter, r *http.Request) {
   // 設置SSE必要的頭信息
   w.Header().Set("Content-Type", "text/event-stream")
   w.Header().Set("Cache-Control", "no-cache")
   w.Header().Set("Connection", "keep-alive")
   w.Header().Set("Access-Control-Allow-Origin", "*")

   // 獲取用戶消息和選擇的模型
   message := r.URL.Query().Get("message")
   model := r.URL.Query().Get("model")

   if message == "" {
      fmt.Println("空消息")
      return
   }

   fmt.Printf("收到消息: %s，使用模型: %s\n", message, model)

   // 創建SSE刷新器
   flusher, ok := w.(http.Flusher)
   if !ok {
      http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
      return
   }
   // 模擬AI回應模式
   responses := simulateAIResponse(message, model)

   // 逐步發送回應片段
   for _, chunk := range responses {
      // 將消息轉換為JSON
      data, err := json.Marshal(chunk)
      if err != nil {
         fmt.Println("JSON編碼錯誤:", err)
         continue
      }

      // 發送SSE格式的消息
      fmt.Fprintf(w, "data: %s\n\n", data)
      flusher.Flush()

      // 模擬打字延遲
      time.Sleep(100 * time.Millisecond)
   }

   // 發送完成信號
   completeMsg := ChatMessage{
      Type:    "complete",
      Content: "",
   }
   completeData, _ := json.Marshal(completeMsg)
   fmt.Fprintf(w, "data: %s\n\n", completeData)
   flusher.Flush()
}

// 處理新對話請求
func handleNewChat(w http.ResponseWriter, r *http.Request) {
   // 返回空的對話區域以重置對話
   w.Header().Set("Content-Type", "text/html")
   fmt.Fprint(w, `<div class="message-container">
      <div class="message ai-message">您好！我已經為您開始了一個新對話。我能幫您什麼忙嗎？</div>
   </div>`)
}

// 處理發送消息的表單提交
func handleSendMessage(w http.ResponseWriter, r *http.Request) {
   // 解析表單
   if err := r.ParseForm(); err != nil {
      fmt.Println(err.Error())
      http.Error(w, "無法解析表單", http.StatusBadRequest)
      return
   }

   // 獲取用戶消息
   message := r.FormValue("user-message")
   if message == "" {
      fmt.Println("empty message")
      http.Error(w, "消息不能為空", http.StatusBadRequest)
      return
   }

   // 回傳用戶消息，實際處理會通過SSE進行
   w.Header().Set("Content-Type", "text/html")
   fmt.Fprintf(w, `<div class="message user-message">%s</div>`, message)
}
