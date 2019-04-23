package main

func main() {
	state := NewLocalState()
	router := MakeServer(state)
	router.Run()
}
