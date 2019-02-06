package lib

import "os"

func PrintErr(message string) {
	_, err := os.Stderr.WriteString(message)
	if err != nil {
		panic(err)
	}
}
