package main

import "github.com/airnomadsmitty/zhuli/zhuli"

func main() {
	zhuli := zhuli.NewZhuli(":8080")
	zhuli.DoTheThing()
}
