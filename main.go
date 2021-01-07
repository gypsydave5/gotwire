package main

import (
	"fmt"
	"github.com/notnil/chess"
	"github.com/notnil/chess/image"
	"html/template"
	"net/http"
)

func templates() *template.Template {
	t := template.Must(template.ParseGlob("./html/*"))
	return t
}

func router() http.Handler {
	game := chess.NewGame()
	var notes []string


	mux := http.NewServeMux()
	t := templates()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		t.ExecuteTemplate(writer, "index.gohtml", "")
	})

	mux.HandleFunc("/chess", func(writer http.ResponseWriter, request *http.Request) {
		t.ExecuteTemplate(writer, "chess.gohtml", notes)
	})

	mux.HandleFunc("/chess_board/move", func(writer http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		move := request.Form.Get("move")
		fmt.Println(move)
		err := game.MoveStr(move)
		if err != nil {
			fmt.Println(err)
		}
		http.Redirect(writer, request, "/chess_board", 301)
	})

	mux.HandleFunc("/notes/new", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			 t.ExecuteTemplate(writer, "new_note.gohtml", "")
			 return
		}
		request.ParseForm()
		note := request.Form.Get("note")
		notes = append(notes, note)
		accept := request.Header.Get("Accept")
		if accept == "text/html; turbo-stream" {
			writer.Header().Add("Content-Type", "text/html; turbo-stream; charset=utf-8")
			fmt.Fprint(writer, `<turbo-stream action="append" target="notes">`)
			fmt.Fprint(writer, `<template>`)
			t.ExecuteTemplate(writer, "note.gohtml", note)
			fmt.Fprint(writer, `</template>`)
			fmt.Fprint(writer, "</turbo-stream>")
		}

	})

	mux.HandleFunc("/chess_board", func(writer http.ResponseWriter, request *http.Request) {
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

	http.ListenAndServe(":9999", router())
}
