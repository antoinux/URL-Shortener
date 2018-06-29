package main

func main() {
	server := NewURLServer(routes)
	server.Start(8080, 8081)
}
