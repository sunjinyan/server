package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Fatal(http.ListenAndServe(":9091", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "hello world")
	//protocol  switch 101 升级协议
	u := &websocket.Upgrader{
		HandshakeTimeout: 0, //握手超时
		ReadBufferSize:   0,
		WriteBufferSize:  0,
		WriteBufferPool:  nil,
		Subprotocols:     nil, //子协议
		Error:            nil,
		CheckOrigin: func(r *http.Request) bool {
			fmt.Println(r.Header.Get("Origin"))
			return true
		}, //检查是否同源，跨域问题
		EnableCompression: false, //是否压缩
	}

	upgrade, err := u.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("cannot upgrade:%v\n", err)
		return
	}
	defer upgrade.Close()

	done := make(chan struct{})

	go func() {
		for {
			m := make(map[string]interface{})
			err = upgrade.ReadJSON(&m)
			if err != nil {
				//fmt.Println(err)
				if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) { //非关闭链接错误，打印错误
					fmt.Printf("unexpected read error : %v\n", err)
				}
				done <- struct{}{}
				break
			}
			fmt.Printf("message received:%v\n", m)
		}
	}()

	i := 0
	for {

		select {
		case <-time.After(3 * time.Second):
		case <-done:
			return
		}

		i++
		err := upgrade.WriteJSON(map[string]string{
			"hello":  "websocket",
			"msg_id": strconv.Itoa(i),
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		//time.Sleep(3*time.Second)
	}
}
