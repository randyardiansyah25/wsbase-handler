# WSBASE-HANDLER

![alt text](https://github.com/randyardiansyah25/wsbase-handler/blob/master/golang.png?raw=true)

Websocket framework base on [gorilla/websocket](https://github.com/gorilla/websocket). Terdapat 2 jenis pengunaan :

1. **websocket server** dapat diimplementasi dengan mudah.
2. **websocket client** yang dapat melakukan koneksi ulang ke server secara otomatis apabila koneksi terputus.

## Example

* [Server example](https://github.com/randyardiansyah25/wsbase-handler/tree/master/sample/server)
* [Client example](https://github.com/randyardiansyah25/wsbase-handler/tree/master/sample/client)

## **Instalasi**

```terminal
go get github.com/randyardiansyah25/wsbase-handler
```

## **Struktur Umum**

```go
type Message struct {
    Type     int         `json:"type"`
    SenderId string      `json:"sender_id"`
    To       []string    `json:"to"`
    Action   string      `json:"action"`
    Title    string      `json:"title"`
    Body     interface{} `json:"body"`
}
```

Merupakan struktur objek pesan yang akan dikirim atau diterima.

Atribut ```Type``` digunakan untuk menandakan pesan broadcast atau pesan private yang bernilai *const* dari ```TypeBroadcast``` atau ```TypePrivate```.

Atribut ```SenderId``` digunakan sebagai informasi identitas pengirim pesan.

Atribut ```To``` berupa objek *```array string```*  digunakan sebagai daftar identitas tujuan pengiriman pesan. Dapat di isi nil atau diabaikan jika ```Type``` bernilan const ```TypeBroadcast```

Atribut ```Action``` dapat digunakan untuk pemetaan penanganan aksi pesan yang akan diolah. Atribut ini disediakan untuk memudahkan pengembang menentukan bisnis proses saat menerima pesan.

Atribut ```Title``` digunakan untuk menyimpan informasi judul dari pesan yang dikirim.

Atribut ```Body``` digunakan untuk menyimpan konten dari pesan yang dikirim.

## **Server**

### **Server Overview**

#### **HUB**

Merupakan peranta untuk melakukan register, unregister dan broadcast pesan dari client. Seluruh client yang terkoneksi, akan terdaftar pada hub dan akan dihapus dalam daftar apabila koneksi terputus. Hub harus dijalankan pertama kali pada aplikasi utama atau method utama.

##### **Initialize**

```go
hub := wsbase.NewHub()
go hub.Run()
```

#### **Const**

```go
const (
    TypeBroadcast = iota
    TypePrivate
)
```

Konstanta ini digunakan untuk menentukan tipe pesan yang akan dikirim. ```TypeBroadcast``` untuk menentukan pesan broadcast yang akan dikirim ke seluruh client terdaftar. ```TypePrivate``` menentukan pesan berupa private yang akan dikirim ke satu atau beberapa client.

#### **Types**

```go
type OnReadMessageFunc func(msg string)
```

Merupakan adapter yang dapat digunakan untuk penangan informasi pesan dikirim oleh client atau yang diterima oleh server.

#### **Method**

```go
func Run()
```

Method ```Run()``` digunakan untuk menjalankan operasi hub. Pemanggilan terhadap method ini bersifat blocking. Dapat dijalankan pada *goroutine* jika ada operasi/statement lanjutan pada aplikasi.

```go
func RegisterClient(id string, w http.ResponseWriter, r *http.Request) error
```

Method ```RegisterClient()``` digunakan untuk mendaftarkan koneksi websocket dari client. Biasa digunakan pada http router.

Parameter ```id string``` digunakan sebagan identitas dari setiap koneksi websocket dari client.

Parameter ```w http.ResponseWriter``` merupakan interface yang digunakan sebagai HTTP handler untuk melakukan kontuksi/membuat respons HTTP.

Parameter ```r *http.Request``` merupakan representasi suatu request HTTP yang diterima oleh server atau yang dikirim oleh client.

```go
func SetOnReadMessageFunc(OnReadMessageFunc)
```

Method ```SetOnReadMessageFunc()``` ini dapat digunakan untuk memodifikasi atau menambahkan handler ketika menerima pesan dari client. Parameter ```OnReadMessageFunc``` harus berupa function yang memiliki parameter dengan tipe data ```string``` yang akan berisi informasi pesan yang dikirim oleh client.

```go
func PushMessage(Message)
```

Method ```PushMessage()``` method yang digunakan untuk melakukan pengiriman pesan dari server. Parameter yang dikirim harus berupa objek dari ```struct Message```

#### Contoh Penggunaan

Contoh implementasi menggunakan http router framework

```go
package main

import (
    "fmt"   
    "github.com/gin-gonic/gin"
    "github.com/randyardiansyah25/wsbase-handler"
)

func main() {
    hub := wsbase.NewHub()
    go hub.Run()    
    r := gin.Default()  
    //accept new client websocket connection
    r.GET("/connect/:id", func(c *gin.Context) {
        hub.RegisterClient(c.Param("id"), c.Writer, c.Request)
    })
    
    //getting request for broadcast message
    r.POST("/broadcast", func(ctx *gin.Context) {
        msg := wsbase.Message{}
        msg.Type = wsbase.TypeBroadcast
        msg.Action = "action_broadcast"
        msg.Title = ctx.Request.FormValue("title")
        msg.Body = ctx.Request.FormValue("message")    
        hub.PushMessage(msg)
        ctx.String(200, "okey")
    })  

    // getting request for send private message
    r.POST("/publish/to", func(ctx *gin.Context) {
        // request payload from another application
        requestPayload := struct {
            Title   string   `form:"title"`
            Message string   `form:"message"`
            To      []string `form:"to"`
        }{}    

        ctx.ShouldBind(&requestPayload)    
        msg := wsbase.Message{}
        msg.Type = wsbase.TypePrivate
        msg.Action = "action_private_message"
        msg.Title = requestPayload.Title
        msg.Body = requestPayload.Message
        msg.To = requestPayload.To
        fmt.Println(requestPayload.To) 
        hub.PushMessage(msg)
        ctx.String(200, "okey")
    })  

    r.Run(":8881")
}
```

## **Client**

### **Client Overview**

##### **Initialize**

```go
client := NewWSClient("addr", "/path/to", false)
client.Start()
```

Parameter pertama, diisi alamat dari websocket server. Parameter kedua diisi dengan path dari endpoint websocket server. Parameter ketiga diisi true jika menggunakan secure websocket (wss), diisi false jika tidak menggunakan secure websocket(ws).


#### **Const**

```go
const (
    LOG = iota
    INFO
    WARN
    ERR
)
```

Merupakan daftar tipe log yang digunakan untuk pemetaan custom log jika digunakan.

#### **Types**

```go
type MessageHandler func(Message)
```

Merupakan adapater yang digunakan untuk penanganan pesan yang diterima dari server. Pengolahan pesan disesuaikan dengan proses bisnis aplikasi setiap pengembang.

```go
type LogHandler func(logType int, val string)
```

Merupakan adapater yang digunakan untuk penanganan log jika digunakan. Secara *default* output yang dicetak ke layar menggunakan *```built in function```* dari framework ini.

#### **Method***

```go
SetMessageHandler(MessageHandler)
```

Method ini dapat digunakan untuk mengirim *handler function*  ke framework. Saat objek *websocke client* menerima pesan, *handler function* ini akan panggil.

```go
SetLogHandler(LogHandler)
```

Method ini digunakan jika akan melakukan kostumisasi log yang dicetak ke layar.

```go
SetReconnectPeriod(time.Duration)
```

Method ini digunakan untuk konfigurasi waktu atau periode dalam melakukan koneksi ulang saat terputus. Contoh ```SetReconnectPeriod(30 * time.Second)``` artinya, saat terjadi putus koneksi, sistem akan melakukan koneksi ulang 30 detik kemudian.

```go
Start() error
```

Method ini digunakan untuk menjalankan *```websocket client```*.

#### Contoh Penggunaan

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"
)

func main() {
    host := "localhost:8881"

    client := NewWSClient(host, "/connect/0001", true)
    client.SetLogHandler(func(errType int, val string) {
        log.Println("ini log custom", val)
    })  

    client.SetReconnectPeriod(5 * time.Second)  
    client.SetMessageHandler(func(m Message) {
        j, _ := json.MarshalIndent(m, "", "    ")
        log.Println("receive message : ")
        fmt.Println(string(j))
    })

    er := client.Start()
    if er != nil {
        log.Println("cannot start ws client, error :", er)
    }
}

```
## Konfigurasi NGINX

```nginx
location /path/to/ {
   proxy_pass http://wsbackend_address;
   proxy_http_version 1.1;
   proxy_set_header Upgrade $http_upgrade;
   proxy_set_header Connection "upgrade";
}
```
