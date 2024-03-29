package main

import (
    "encoding/json"
    "bytes"
    "strconv"
    "time"
    "fmt"
    "log"
    "os"
    "flag"
    "strings"
    "net/http"
    "github.com/gorilla/sessions"
    "github.com/gorilla/websocket"
)


var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

var EASYBITS_URL string
var BEARER_TOKEN string

var user_connections = map[string]*websocket.Conn{}

var (
    // key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
    key = []byte("super-secret-key")
    store = sessions.NewCookieStore(key)
    chat_logs = map[string]ChatData{}
)

type Message struct {
    User string
    Text string
    Type string
}

type ChatData struct {
    Messages []Message
}

type UserInfo struct {
  Id string `json:"id"`
}

type IncomingMessage struct {
  Recipient UserInfo `json:"recipient"`
  Type string `json:"type"`
  Data string `json:"data"`
  MimeType string `json:"mimeType"`
}

type BackendMessage struct {
  Message IncomingMessage `json:"message"`
  Timestamp string `json:"timestamp"`
}

func performRequest(url string, body []byte) {
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Authorization", "Bearer " + BEARER_TOKEN)
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
}

func getPathValue(path string, position int) string {
  var pathValues = strings.Split(path, "/")
  return pathValues[position + 1]
}

func chat(w http.ResponseWriter, req *http.Request) {
    conn, _ := upgrader.Upgrade(w, req, nil) // error ignored for sake of simplicity

    // create random user id
    userId := getPathValue(req.URL.Path, 1)
    user_connections[userId] = conn

    for _, chat_data := range chat_logs[userId].Messages {
      if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"text": "`+chat_data.Text+`", "user": "`+chat_data.User+`", "type": "`+chat_data.Type+`"}`)); err != nil {
          return
      }
    }

    for {
        // Read message from browser
        msgType, msg, err := conn.ReadMessage()
        if err != nil {
            return
        }

        // Print the message to the console
        var text = string(msg)
        // perform request to backend
        var timestamp = strconv.FormatInt(time.Now().Unix(), 10)
        performRequest(EASYBITS_URL, []byte(`{"message": {"recipient": {"id": "`+string(userId)+`"}, "text": "`+text+`"}, "timestamp": "`+string(timestamp)+`"}`))

        chat_data := chat_logs[userId]
        messages := append(chat_data.Messages, Message{Text: text, User: "User", Type: "text"})
        chat_logs[userId] = ChatData{Messages: messages}

        // Write message back to browser
        if err = conn.WriteMessage(msgType, []byte(`{"text": "`+string(msg)+`", "user": "User", "type": "text"}`)); err != nil {
            return
        }
    }
}

func chat_window(w http.ResponseWriter, req *http.Request) {
  http.ServeFile(w, req, "templates/layout.html")
}

func chat_backend(w http.ResponseWriter, req *http.Request) {
  var message BackendMessage
  json.NewDecoder(req.Body).Decode(&message)

  var userId = message.Message.Recipient.Id
  if _, ok := chat_logs[userId]; !ok {
    chat_logs[userId] = ChatData{
      Messages: []Message{
        {Text: message.Message.Data, User: "Bot",},
      },
    }
  } else {
    var msgs = append(chat_logs[userId].Messages, Message{Text: message.Message.Data, User: "Bot", Type: message.Message.Type})
    chat_logs[userId] = ChatData{Messages: msgs}
  }
  w.WriteHeader(http.StatusOK)

  user_conn := user_connections[userId]
  if user_conn != nil {
    user_conn.WriteMessage(websocket.TextMessage, []byte(`{"text": "`+message.Message.Data+`", "user": "Bot", "type": "`+message.Message.Type+`"}`))
  }

  return
}

func usage() {
    fmt.Fprintf(os.Stderr, "usage: %s [inputfile]\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(2)
}

func main() {

    flag.Usage = usage
    portPtr := flag.Int("port", 8090, "an int")
    urlPtr := flag.String("easybits_url", "https://api-controller.easybits.tech/api/<your_id>", "The url to the easybits api")
    tokenPtr := flag.String("bearer", "<your_token>", "The bearer token for the easybits api")
    flag.Parse()

    EASYBITS_URL = *urlPtr
    BEARER_TOKEN = *tokenPtr

    http.HandleFunc("/backend/", chat_backend)
    http.HandleFunc("/chat/{chat_id}/send", chat)
    http.HandleFunc("GET /chat/{chat_id}", chat_window)

    err := http.ListenAndServe(":" + strconv.FormatInt(int64(*portPtr), 10), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
