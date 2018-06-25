package main

func main() {
	server := NewRestServer()
	server.Start(8080)
}
