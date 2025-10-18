package main

func main() {
	if err := DownloadChainRegistry(); err != nil {
		panic(err)
	}
}