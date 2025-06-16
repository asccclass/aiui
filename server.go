package main

import (
   "os"
   "fmt"
   "github.com/joho/godotenv"
   "github.com/asccclass/sherryserver"
)

var AIs map[string]interface{}
var McpClient *MCPClient       // MCPClient 用於與 MCP Server 進行交互

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
   port := os.Getenv("PORT")
   if port == "" {
      port = "80"
   }
   documentRoot := os.Getenv("DocumentRoot")
   if documentRoot == "" {
      documentRoot = "www"
   }
   templateRoot := os.Getenv("TemplateRoot")
   if templateRoot == "" {
      templateRoot = "www/html"
   }

   server, err := SherryServer.NewServer(":" + port, documentRoot, templateRoot)
   if err != nil {
      panic(err)
   }
   AIs = make(map[string]interface{})
   router := NewRouter(server, documentRoot)
   if router == nil {
      fmt.Println("router return nil")
      return
   }
   // 啟動 MCP 客戶端
   McpClient = NewMCPClient()
   if McpClient == nil {   
      fmt.Println("MCPClient is nil")
      return
   }
   if err := McpClient.Connect(); err != nil { // 連接到 MCP Server
      fmt.Printf("連接到 MCP Server 失敗: %s\n", err.Error())
      return
   }
   
   // 新版 MCP Host   
   mcpHost := NewMCPHost()
   serviceName := "todo" // 這裡可以替換為實際的服務名稱
   endpoint := "https://www.justdrink.com.tw/mcpsrv/capabilities/" + serviceName
   if err := mcpHost.GetCapabilities(serviceName, endpoint); err != nil {
      fmt.Printf("獲取 MCP Server 能力失敗: %s\n", err.Error())   
      return
   }

   server.Server.Handler = router  // server.CheckCROS(router)  // 需要自行implement, overwrite 預設的
   server.Start()
}
