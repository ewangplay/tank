package main

import (
	"bytes"
	"io"
	"log"
	"os"
)

// 碰撞检测函数
func checkCollision(x1, y1, w1, h1, x2, y2, w2, h2 float32) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

func loadFromFile(filePath string) io.Reader {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fontData, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return bytes.NewReader(fontData)
}
