package main

import "github/com/kawakami-o3/gut"

func main() {
	client := gut.Client{
		Address: "127.0.0.1:9000",
	}

	client.Run()
}
