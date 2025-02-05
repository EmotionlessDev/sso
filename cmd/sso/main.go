package main

import (
	"fmt"

	"github.com/EmotionlessDev/sso/internal/config"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)
}
