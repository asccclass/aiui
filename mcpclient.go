package main

import(
   "fmt"
   "strconv"
)
/*
return map[string]interface{} {
	"is_todo_related": false,
	"action":          "general_chat",
	"parameters":      map[string]interface{}{},
}
*/         
// 執行 Tools 工具
func RunTools(req GenerateRequest, prompt string)(string, error) {
   s, err := parseIntent(req, prompt) // (map[string]interface{}, error)	
   if err != nil {
      return "", err
   }
   isTodoRelated, ok := s["is_todo_related"].(bool)
   if !ok || !isTodoRelated {
      return "", fmt.Errorf("not related tools")
   }
   action, ok := s["action"].(string)
   if !ok {
	   return "", fmt.Errorf("No action found in intent")
	}
   parameters, ok := s["parameters"].(map[string]interface{})
	if !ok {
		parameters = make(map[string]interface{})
	}
   
	switch action {  // 根據動作調用相應的 MCP 工具
	case "get_all":
		res, err := callMCPTool("get_all_todos", make(map[string]interface{}))
		if err != nil {
			fmt.Printf("調用 get_all_todos 工具失敗: %s", err.Error())
		}
		return res, nil
	case "get_by_id":
		if idVal, exists := parameters["id"]; exists {
			var id int
			switch v := idVal.(type) {
			case float64:
				id = int(v)
			case int:
				id = v
			case string:
				if parsed, err := strconv.Atoi(v); err == nil {
					id = parsed
				} else {
					return "", fmt.Errorf("無法解析待辦事項 ID")
				}
			default:
				return "", fmt.Errorf("無效的待辦事項 ID 格式")
			}
			return callMCPTool("get_todo_by_id", map[string]interface{}{"id": id})
		}
		return "", fmt.Errorf("請提供待辦事項的 ID")

	case "create":
		context, hasContext := parameters["context"].(string)
		user, hasUser := parameters["user"].(string)
		if !hasContext || context == "" {
			context = prompt  // 嘗試從原始輸入中提取內容
		}
		if !hasUser || user == "" {
			user = "預設使用者"
		}
		args := map[string]interface{}{
			"context": context,
			"user":    user,
		}
		// 添加可選參數
		if dueTime, exists := parameters["duetime"]; exists {
			args["duetime"] = dueTime
		}
		if isFinish, exists := parameters["isFinish"]; exists {
			args["isFinish"] = isFinish
		}

		return callMCPTool("create_todo", args)

	case "update":
		if idVal, exists := parameters["id"]; exists {
			var id int
			switch v := idVal.(type) {
			case float64:
				id = int(v)
			case int:
				id = v
			case string:
				if parsed, err := strconv.Atoi(v); err == nil {
					id = parsed
				} else {
					return "", fmt.Errorf("無法解析待辦事項 ID")
				}
			default:
				return "", fmt.Errorf("無效的待辦事項 ID 格式")
			}
			args := map[string]interface{}{"id": id}
	
			// 添加其他更新參數
			for key, value := range parameters {
				if key != "id" {
					args[key] = value
				}
			}
			return callMCPTool("update_todo", args)
		}
		return "", fmt.Errorf("請提供要更新的待辦事項 ID")

	case "delete":
		if idVal, exists := parameters["id"]; exists {
			var id int
			switch v := idVal.(type) {
			case float64:
				id = int(v)
			case int:
				id = v
			case string:
				if parsed, err := strconv.Atoi(v); err == nil {
					id = parsed
				} else {
					return "", fmt.Errorf("無法解析待辦事項 ID")
				}
			default:
				return "", fmt.Errorf("無效的待辦事項 ID 格式")
			}
			return callMCPTool("delete_todo", map[string]interface{}{"id": id})
		}
		return "", fmt.Errorf("請提供要刪除的待辦事項 ID")

	default:  // 未知動作，使用一般對話		
		return "", fmt.Errorf("未知的動作類型: %s", action)
	}
   return prompt, nil
}
