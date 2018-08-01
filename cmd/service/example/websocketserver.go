package example

import (
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"html/template"
	"log"
	// "net/http"
	// "net/url"
)

var upgrader = websocket.Upgrader{}

func Echo(ctx echo.Context) error {

	// host := "locahost:9003"
	// var u = url.URL{Scheme: "http", Host: host, Path: "/echo"}
	conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		log.Print("upgrade: ", err)
		return err
	}

	defer conn.Close()

	for {
		qt, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("read: ", err)
			break
		}

		log.Printf("recv: %s", string(data))

		err = conn.WriteMessage(qt, data)
		if err != nil {
			log.Println("write: ", err)
			break
		}
	}
	return nil
}

func Home(ctx echo.Context) error {
	return homeTemplate.Execute(ctx.Response(), "ws://"+ctx.Request().Host+"/example/echo")
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
