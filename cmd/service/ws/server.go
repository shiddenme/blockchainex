package ws

import (
	"bufio"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	writeWait        = 10 * time.Second
	maxMessageSize   = 8192
	pongWait         = 60 * time.Second
	pingPeriod       = (pongWait * 9) / 10
	closeGradePeriod = 10 * time.Second
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func pumpStdin(conn *websocket.Conn, w io.Writer) {
	defer conn.Close()
	conn.SetReadLimit(int64(maxMessageSize))
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		message = append(message, '\n')
		if _, err := w.Write(message); err != nil {
			break
		}
	}
}

func pumpStuout(conn *websocket.Conn, r io.Reader, done chan struct{}) {
	defer func() {}()
	s := bufio.NewScanner(r)
	for s.Scan() {
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteMessage(websocket.TextMessage, s.Bytes()); err != nil {
			conn.Close()
			break
		}
	}
	if s.Err() != nil {
		log.Fatalln("scan: ", s.Err())
	}
	close(done)

	conn.SetWriteDeadline(time.Now().Add(writeWait))
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(closeGradePeriod)
	conn.Close()
}

func ping(conn *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Fatalln("ping: ", err)
			}
		case <-done:
			return
		}
	}
}

func internalError(conn *websocket.Conn, msg string, err error) {
	log.Println(msg, err)
	conn.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
}

func ServeWs(ctx echo.Context) error {
	conn, err := wsUpgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		log.Fatalln("upgrade fail, ", err)
		return err
	}

	defer conn.Close()

	outr, outw, err := os.Pipe()
	if err != nil {
		log.Fatalln("stdout: ", err)
		return err
	}

	defer outr.Close()
	defer outw.Close()

	inr, inw, err := os.Pipe()
	if err != nil {
		log.Fatalln("stdin: ", err)
		return err
	}
	defer inr.Close()
	defer inw.Close()

	proc, err := os.StartProcess("", flag.Args(), &os.ProcAttr{
		Files: []*os.File{inr, outw, outw},
	})
	if err != nil {
		log.Fatalln("start: ", err)
		return err
	}
	inr.Close()
	outw.Close()

	stdoutDone := make(chan struct{})

	go pumpStuout(conn, outr, stdoutDone)
	go ping(conn, stdoutDone)

	pumpStdin(conn, inw)

	if err := proc.Signal(os.Interrupt); err != nil {
		log.Fatalln("inter: ", err)
	}

	select {
	case <-stdoutDone:
	case <-time.After(time.Second):
		if err := proc.Signal(os.Kill); err != nil {
			log.Fatalln("term: ", err)
		}
		<-stdoutDone
	}

	if _, err := proc.Wait(); err != nil {
		log.Fatalln("wait: ", err)
	}
	return nil
}
