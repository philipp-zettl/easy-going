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

func escape_string(text string) string {
  return strings.Replace(
    strings.Replace(
      strings.Replace(
        strings.Replace(
          strings.Replace(
            strings.Replace(
              strings.Replace(
                string(text), "\\", "\\\\", -1,
              ), "\\n", "\\\\n", -1,
            ), "\\r", "\\\\r", -1,
          ), "\\\"", "\\\\\"", -1,
        ), "\\'", "\\\\'", -1,
      ), "\\`", "\\\\`", -1,
    ), "\\t", "\\\\t", -1,
  )
}

func chat(w http.ResponseWriter, req *http.Request) {

    var isBackend = strings.Contains(req.URL.Path, "/backend")
    if isBackend {
      chat_backend(w, req)
      return
    }
    var isChat = strings.Contains(req.URL.Path, "/chat")
    if !isChat {
      return
    }
    var isSend = strings.Contains(req.URL.Path, "/send")

    if !isSend {
      http.ServeFile(w, req, "templates/layout.html")
      return
    }

    conn, _ := upgrader.Upgrade(w, req, nil) // error ignored for sake of simplicity

    // create random user id
    userId := getPathValue(req.URL.Path, 1)
    user_connections[userId] = conn

    for _, chat_data := range chat_logs[userId].Messages {
      b, err := json.Marshal(chat_data)
      if err != nil {
          fmt.Println(err)
          return
      }

      if err := conn.WriteMessage(websocket.TextMessage, []byte(string(b))); err != nil {
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
        var text = escape_string(string(msg))
        if text == "[CLEAR]" {
          chat_logs[userId] = ChatData{Messages: []Message{}}
          continue
        }
        // perform request to backend
        var timestamp = strconv.FormatInt(time.Now().Unix(), 10)
        performRequest(EASYBITS_URL, []byte(`{"message": {"recipient": {"id": "`+string(userId)+`"}, "text": "`+escape_string(text)+`"}, "timestamp": "`+string(timestamp)+`"}`))

        chat_data := chat_logs[userId]
        new_message := Message{Text: text, User: "User", Type: "text"}
        messages := append(chat_data.Messages, new_message)
        chat_logs[userId] = ChatData{Messages: messages}
        b, err := json.Marshal(new_message)
        if err != nil {
            fmt.Println(err)
            return
        }


        // Write message back to browser
        if err = conn.WriteMessage(msgType, []byte(string(b))); err != nil {
            return
        }
    }
}

func chat_backend(w http.ResponseWriter, req *http.Request) {
  var message BackendMessage
  json.NewDecoder(req.Body).Decode(&message)

  var userId = message.Message.Recipient.Id
  new_message := Message{Text: message.Message.Data, User: "Bot", Type: message.Message.Type}

  if _, ok := chat_logs[userId]; !ok {
    chat_logs[userId] = ChatData{
      Messages: []Message{
        new_message,
      },
    }
  } else {
    var msgs = append(chat_logs[userId].Messages, new_message)
    chat_logs[userId] = ChatData{Messages: msgs}
  }
  w.WriteHeader(http.StatusOK)

  user_conn := user_connections[userId]
  if user_conn != nil {
    b, err := json.Marshal(new_message)
    if err != nil {
        fmt.Println(err)
        return
    }

    user_conn.WriteMessage(websocket.TextMessage, []byte(string(b)));
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

    http.HandleFunc("/", chat)

    log.Println("Server starting on port " + strconv.FormatInt(int64(*portPtr), 10))
    err := http.ListenAndServe(":" + strconv.FormatInt(int64(*portPtr), 10), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
