package main

import (
   // "os"
   "fmt"
   // "time"
   "strings"
   "encoding/json"
)

// parseIntentWithOllama 使用 Ollama 解析使用者意圖
func parseIntent(userInput string) (map[string]interface{}, error) {
   prompt := fmt.Sprintf(`你是一個待辦事項助手。請分析以下使用者輸入，判斷是否與待辦事項相關。

使用者輸入："%s"

請回應一個 JSON 格式，包含以下欄位：
- "is_todo_related": true/false (是否與待辦事項相關)
- "action": "get_all" | "get_by_id" | "create" | "update" | "delete" | "general_chat" (動作類型)
- "parameters": {} (相關參數，如果有的話)

如果是待辦事項相關：
- 查看/列出待辦事項 -> action: "get_all"
- 查看特定待辦事項 -> action: "get_by_id", parameters: {"id": 數字}
- 新增/建立待辦事項 -> action: "create", parameters: {"context": "內容", "user": "使用者"}
- 修改/更新待辦事項 -> action: "update", parameters: {"id": 數字, 其他欄位...}
- 刪除待辦事項 -> action: "delete", parameters: {"id": 數字}

如果不是待辦事項相關 -> action: "general_chat"

請只回應 JSON，不要其他文字。`, userInput)

   response := ""
   var err error
   if AIs["ollama"] != nil {
      response, err := AIs["Ollama"].(Ollama).Send2LLM(prompt)  // (string, error) 
      if err != nil {
         return nil, fmt.Errorf("query ollama for intent: %s", err.Error())
      }
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