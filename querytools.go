package main

import (
   "fmt"
   "strings"
   "encoding/json"
)

// parseIntentWithOllama 使用 Ollama 解析使用者意圖
func parseIntent(req GenerateRequest, userInput string) (map[string]interface{}, error) {
   prompt := fmt.Sprintf(`你是一個待辦事項助手。請分析以下使用者輸入，判斷是否與待辦事項相關。

請依據使用者輸入，回應一個 JSON 格式，包含以下欄位：
- "is_todo_related": true/false (是否與待辦事項相關)
- "action": "get_all" | "get_by_id" | "create" | "update" | "delete" | "general_chat" (動作類型)
- "parameters": {} (相關參數，如果有的話)

如果是待辦事項相關：
- 查看/列出待辦事項 -> action: "get_all"
- 查看特定待辦事項 -> action: "get_by_id", parameters: {"id": "string"}
- 新增/建立待辦事項 -> action: "create", parameters: {"context": "內容", "user": "負責者", "isFinished": "0"}
- 修改/更新待辦事項 -> action: "update", parameters: {"id": "文字", 其他欄位...}
- 刪除待辦事項 -> action: "delete", parameters: {"id": "string"}
- 若欄位名稱為status/狀態時，請將欄位名稱替換為isFinish欄位：
   - 當status/狀態為「未處理」、「未分類」、「未知」或類似狀態時，isFinish應設為 "0"
   - 當status/狀態為「進行中」、「待處理」、「審核中」、「處理中/open/process/in progress」或類似狀態時，isFinish應設為 "1"
   - 當status/狀態為「完成/completed/done」、「已結束/closed」、「已處理/finished」或類似狀態時，isFinish應設為 "2"
   - 當status/狀態為「擱置/rejected」、「不處理」、「暫存/pause/suspended/pending」或類似狀態時，isFinish應設為 "3"
- 當使用者輸入的文字中包含「時間相關語句」（如「明天上午十點」、「下週三下午三點」等），請將該時間轉換為標準 UTC 格式，格式為 YYYY-MM-DD HH:MM:SS。
   - 例如使用者輸入「明天上午十點開會」，請判斷當前時間(如：2025-06-14）為基準，加一天後時間為 2025-06-15 10:00:00，並將此值存為欄位名稱：duetime。
   - 若時間資訊模糊（如「晚上」、「中午」），請估算合理時間（如「晚上」可設為 20:00:00）。
   - 若沒有指定年份、月份、日期，則年月日為當前西元年月日。
   - 輸出要求：
      - 時間應為完整的 UTC 格式時間字串（YYYY-MM-DD HH:MM:SS）。
      - 資料應儲存在欄位 duetime。

如果不是待辦事項相關 -> action: "general_chat"
請只回應 JSON，不要其他文字。

使用者輸入："%s"
`, userInput)

   response := ""
   var err error
   if AIs["Ollama"].(*OllamaClient) != nil {
      jData, err := AIs["Ollama"].(*OllamaClient).Prompt2String(req, "user", prompt)
      if err != nil {
         return nil, fmt.Errorf("prepare prompt for ollama: %s", err.Error())
      }
      res, err := AIs["Ollama"].(*OllamaClient).Send2LLM(jData)  // (string, error) 
      if err != nil {
         return nil, fmt.Errorf("query ollama for intent: %s", err.Error())
      }
      response = res
   } else {
      fmt.Println("No Ollama client initialized, cannot parse intent")
      return nil, fmt.Errorf("Not any LLM is initialized")
   }
   // 清理回應，只保留 JSON 部分
   response = strings.TrimSpace(response)
   start := strings.Index(response, "{")
   end := strings.LastIndex(response, "}") + 1
   if start >= 0 && end > start {
      response = response[start:end]
   }
   var intent map[string]interface{}
   if err = json.Unmarshal([]byte(response), &intent); err != nil { // 如果 JSON 解析失敗，預設為一般對話
      return map[string]interface{}{
         "is_todo_related": false,
         "action":          "general_chat",
         "parameters":      map[string]interface{}{},
      }, nil
   }   
   return intent, nil
}