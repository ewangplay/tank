package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// 屏幕宽高
	screenWidth = 640
	// 屏幕高度
	screenHeight = 480
	// 坦克的速度
	tankSpeed = 2
	// 子弹速度
	bulletSpeed = 5
	// 每 30 帧改变一次方向
	changeDirInterval = 30
	// 每 5 帧射击一次
	shootInterval = 5
	// 最大敌方坦克数量
	maxEnemyTankCount = 5
	// 最大墙的数量
	maxWallCount = 5
	// 墙的检测间隔，分钟为单位
	wallCheckInterval = 1
	// 敌方坦克的检测间隔，分钟为单位
	enemyTankCheckInterval = 2
	// 状态栏的高度
	statusBarHeight = 20
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Tank Game")
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Game Over")
}
