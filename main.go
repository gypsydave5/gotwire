package main

import (
	"context"
	"fmt"
	"github.com/notnil/chess"
	"github.com/notnil/chess/image"

	"html/template"
	"net/http"
	"nhooyr.io/websocket"
	"os"
)

func templates() *template.Template {
	t := template.Must(template.ParseGlob("./html/*"))
	return t
}

func router() http.Handler {
	game := chess.NewGame()
	//var notes []string

	// WORLD'S DUMBEST PUB/SUB TM
	var conns []*websocket.Conn

	mux := http.NewServeMux()
	t := templates()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		t.ExecuteTemplate(writer, "index.gohtml", "")
	})

	mux.HandleFunc("/chess", func(writer http.ResponseWriter, request *http.Request) {
		t.ExecuteTemplate(writer, "chess.gohtml", "")
	})

	// Ws handler
	mux.HandleFunc("/chess_board/ws", func(writer http.ResponseWriter, request *http.Request) {
		conn, err := websocket.Accept(writer, request, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		conns = append(conns, conn)
		// TODO: remove conn when connection is broken? Or find a decent ws pub sub lib.
	})

	mux.HandleFunc("/chess_board/move", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("MOVE")
		request.ParseMultipartForm(256)

		for key, value := range request.PostForm {
			fmt.Println(key, value)
		}
		move := request.PostForm.Get("move")
		fmt.Println(move)
		err := game.MoveStr(move)
		if err != nil {
			fmt.Println(err)
		}
		// for each ws conn registered, send the turbo stream with the updated board
		// todo: make the board turbo-frame into a gohtml template to deuglify this
		for _, conn := range conns {
			wsWriter, _ := conn.Writer(context.TODO(), websocket.MessageText)
			defer wsWriter.Close()
			fmt.Fprint(wsWriter, `{ 
				"identifier": 
				   "{\"channel\":\"Turbo::StreamsChannel\",\"signed_stream_name\":\"**mysignature**\"}",
				"message":
				  "<turbo-stream action="replace" target="chess_board">
					<template>
<turbo-frame id="chess_board" src="/chess_board">`)
			image.SVG(wsWriter, game.Position().Board())
			fmt.Fprint(wsWriter, `
</turbo-frame>
</template>
				   </turbo-stream>"
			  }`)
		}

		// http.Redirect(writer, request, "/chess_board", 303)
	})

	//mux.HandleFunc("/notes/new", func(writer http.ResponseWriter, request *http.Request) {
	//	if request.Method != http.MethodPost {
	//		 t.ExecuteTemplate(writer, "new_note.gohtml", "")
	//		 return
	//	}
	//	request.ParseForm()
	//	note := request.Form.Get("note")
	//	notes = append(notes, note)
	//	accept := request.Header.Get("Accept")
	//	if strings.Contains(accept, "text/html; turbo-stream") {
	//		writer.Header().Add("Content-Type", "text/html; turbo-stream; charset=utf-8")
	//		fmt.Fprint(writer, `<turbo-stream action="append" target="notes">`)
	//		fmt.Fprint(writer, `<template>`)
	//		t.ExecuteTemplate(writer, "note.gohtml", note)
	//		fmt.Fprint(writer, `</template>`)
	//		fmt.Fprint(writer, "</turbo-stream>")
	//	}
	//
	//})

	// todo: this is duplicated a lot
	mux.HandleFunc("/chess_board", func(writer http.ResponseWriter, request *http.Request) {
		accept := request.Header.Get("Accept")
		fmt.Println(accept)
		fmt.Fprintln(writer, `<html><body>`)
		fmt.Fprintln(writer, `<h1>It's a Chess Board</h1>'`)
		fmt.Fprintln(writer, `<turbo-frame id="chess_board">`)
		image.SVG(writer, game.Position().Board())
		fmt.Fprintln(writer, `<turbo-frame />`)
		fmt.Fprintln(writer, `</html></body>`)
	})
	return mux
}

func main() {
	fmt.Println("hello")
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "9999"
	}

	fmt.Printf("Listening on port %s\n", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), router())
}
