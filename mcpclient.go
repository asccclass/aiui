package main

import(
   "fmt"
)

// 執行 Tools 工具
func RunTools(p string)(map[string]interface, error) {
   s, err := parseIntent(p) // (map[string]interface{}, error)
   if err != nil {
      return nil, err
   }
   if !s["is_todo_related"] {
      return nil, fmt.Errorf("not related tools")
   }
   return s, nil
}
