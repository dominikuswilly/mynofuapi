package main
import "os"
func main() {
    os.WriteFile("hello_debug.txt", []byte("Hello!"), 0644)
}
