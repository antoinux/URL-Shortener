package main

func main() {
	server := NewRestServer(routes)
	server.Start(8080)
}
