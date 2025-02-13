package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	mplusFaceSource *text.GoTextFaceSource
)

func init() {
	s, err := text.NewGoTextFaceSource(loadFromFile("fonts/STSONG.ttf"))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s
}

// Tank 表示坦克
type Tank struct {
	x, y             float32
	direction        int // 0: 上, 1: 右, 2: 下, 3: 左
	directionChanged bool
	hasShot          bool
	health           int
}

// Bullet 表示子弹
type Bullet struct {
	x, y      float32
	direction int
}

// Wall 表示墙壁
type Wall struct {
	x, y          float32
	width, height float32
	health        int
}

// Game 表示游戏状态
type Game struct {
	playerTank     *Tank
	bossTank       *Tank
	playerBullets  []Bullet
	enemyTanks     []Tank
	bossBullets    []Bullet
	enemyBullets   []Bullet
	walls          []Wall
	enemyTimer     *time.Timer
	wallTimer      *time.Timer
	gameOver       bool
	gameSucc       bool
	enemyTankCount int
	lastFollowTime time.Time
}

// NewGame 创建一个新的游戏实例
func NewGame() *Game {
	game := &Game{
		playerTank: &Tank{
			x:         screenWidth / 2,
			y:         screenHeight/2 + statusBarHeight,
			direction: 0,
			health:    playerTankHP,
		},
		playerBullets: []Bullet{},
		bossTank: &Tank{
			x:         100,
			y:         100,
			direction: 2,
			health:    bossTankHP,
		},
		enemyTanks:   []Tank{},
		enemyBullets: []Bullet{},
		bossBullets:  []Bullet{},
		walls: []Wall{
			{x: 150, y: 150, width: 100, height: 10, health: wallHP},
			{x: 250, y: 280, width: 150, height: 10, health: wallHP},
			{x: 400, y: 50, width: 10, height: 100, health: wallHP},
			{x: 350, y: 350, width: 10, height: 50, health: wallHP},
		},
		enemyTimer:     time.NewTimer(time.Duration(5+rand.Intn(enemyTankCheckInterval)) * time.Second),
		wallTimer:      time.NewTimer(time.Duration(10+rand.Intn(wallCheckInterval)) * time.Second),
		gameOver:       false,
		gameSucc:       false,
		enemyTankCount: maxEnemyTankCount,
		lastFollowTime: time.Now(),
	}

	go game.spawnEnemyTanks()
	go game.spawnWalls()

	return game
}

// spawnEnemyTanks 定时生成敌方坦克
func (g *Game) spawnEnemyTanks() {
	for {
		<-g.enemyTimer.C
		if len(g.enemyTanks) < g.enemyTankCount {
			newTank := Tank{
				x:         float32(rand.Intn(screenWidth - 20)),
				y:         float32(statusBarHeight + rand.Intn(screenHeight-40)),
				direction: rand.Intn(4),
				health:    enemyTankHP,
			}
			g.enemyTanks = append(g.enemyTanks, newTank)
		}
		g.enemyTimer.Reset(time.Duration(5+rand.Intn(enemyTankCheckInterval)) * time.Second)
	}
}

// spawnWalls 定时生成墙
func (g *Game) spawnWalls() {
	for {
		<-g.wallTimer.C
		if len(g.walls) < maxWallCount {
			var newWall Wall
			if rand.Intn(2) == 0 {
				// 生成水平的墙
				newWall = Wall{
					x:      float32(rand.Intn(screenWidth - 50)),
					y:      float32(statusBarHeight*2 + rand.Intn(screenHeight-10)),
					width:  float32(rand.Intn(50) + 50),
					height: 10,
					health: wallHP,
				}
			} else {
				// 生成竖直的墙
				newWall = Wall{
					x:      float32(rand.Intn(screenWidth - 10)),
					y:      float32(statusBarHeight + rand.Intn(screenHeight-50)),
					width:  10,
					height: float32(rand.Intn(50) + 50),
					health: wallHP,
				}
			}
			g.walls = append(g.walls, newWall)
		}
		g.wallTimer.Reset(time.Duration(10+rand.Intn(wallCheckInterval)) * time.Second)
	}
}

func (g *Game) updatePlayerTank() error {
	if g.playerTank != nil {
		// 处理坦克移动
		var newX, newY = g.playerTank.x, g.playerTank.y

		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.playerTank.direction = 0
			if g.playerTank.y > statusBarHeight {
				newY -= tankSpeed
			}
		} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
			g.playerTank.direction = 1
			if g.playerTank.x < screenWidth-20 {
				newX += tankSpeed
			}
		} else if ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.playerTank.direction = 2
			if g.playerTank.y < screenHeight-20 {
				newY += tankSpeed
			}
		} else if ebiten.IsKeyPressed(ebiten.KeyLeft) {
			g.playerTank.direction = 3
			if g.playerTank.x > 0 {
				newX -= tankSpeed
			}
		}

		collision := false
		// 检测与Boss坦克的碰撞
		if g.bossTank != nil {
			if checkCollision(newX, newY, 20, 20, g.bossTank.x, g.bossTank.y, 20, 20) {
				collision = true
			}
		}

		// 检测与敌方坦克的碰撞
		for _, enemyTank := range g.enemyTanks {
			if checkCollision(newX, newY, 20, 20, enemyTank.x, enemyTank.y, 20, 20) {
				collision = true
				break
			}
		}

		// 检测与墙的碰撞
		for _, wall := range g.walls {
			if checkCollision(newX, newY, 20, 20, wall.x, wall.y, wall.width, wall.height) {
				collision = true
				break
			}
		}

		// 如果没有碰撞，更新坦克位置
		if !collision {
			g.playerTank.x = newX
			g.playerTank.y = newY
		}

		// 处理射击
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			bullet := Bullet{
				x:         g.playerTank.x + 8,
				y:         g.playerTank.y + 8,
				direction: g.playerTank.direction,
			}
			g.playerBullets = append(g.playerBullets, bullet)
		}
	}
	return nil
}

func (g *Game) updatePlayerBullets() error {

	// 更新子弹位置
	for i := 0; i < len(g.playerBullets); i++ {
		switch g.playerBullets[i].direction {
		case 0:
			g.playerBullets[i].y -= bulletSpeed
		case 1:
			g.playerBullets[i].x += bulletSpeed
		case 2:
			g.playerBullets[i].y += bulletSpeed
		case 3:
			g.playerBullets[i].x -= bulletSpeed
		}

		// 检测玩家子弹与Boss坦克的碰撞
		if g.bossTank != nil {
			if checkCollision(g.playerBullets[i].x, g.playerBullets[i].y, 5, 5, g.bossTank.x, g.bossTank.y, 20, 20) {
				g.bossTank.health--
				if g.bossTank.health <= 0 {
					// 移除Boss坦克
					g.bossTank = nil
				}
				// 移除子弹
				g.playerBullets = append(g.playerBullets[:i], g.playerBullets[i+1:]...)
				i--
			}
		}

		// 检测玩家子弹与敌方坦克的碰撞
		for j := 0; j < len(g.enemyTanks); j++ {
			if i < 0 {
				break
			}
			if checkCollision(g.playerBullets[i].x, g.playerBullets[i].y, 5, 5, g.enemyTanks[j].x, g.enemyTanks[j].y, 20, 20) {
				g.enemyTanks[j].health--
				if g.enemyTanks[j].health <= 0 {
					// 移除敌方坦克
					g.enemyTanks = append(g.enemyTanks[:j], g.enemyTanks[j+1:]...)
					g.playerTank.health++
				}
				// 移除子弹
				g.playerBullets = append(g.playerBullets[:i], g.playerBullets[i+1:]...)
				i--
				break
			}
		}

		// 检测玩家子弹与墙的碰撞
		for j := 0; j < len(g.walls); j++ {
			if i < 0 {
				break
			}
			if checkCollision(g.playerBullets[i].x, g.playerBullets[i].y, 5, 5, g.walls[j].x, g.walls[j].y, g.walls[j].width, g.walls[j].height) {
				g.walls[j].health--
				if g.walls[j].health <= 0 {
					// 移除墙
					g.walls = append(g.walls[:j], g.walls[j+1:]...)
				}
				// 移除子弹
				g.playerBullets = append(g.playerBullets[:i], g.playerBullets[i+1:]...)
				i--
				break
			}
		}

		// // 检测玩家子弹与Boss子弹的碰撞
		// for j := 0; j < len(g.bossBullets); j++ {
		// 	if i < 0 {
		// 		break
		// 	}
		// 	if checkCollision(g.playerBullets[i].x, g.playerBullets[i].y, 5, 5, g.bossBullets[j].x, g.bossBullets[j].y, 5, 5) {
		// 		// 移除Boss子弹
		// 		g.bossBullets = append(g.bossBullets[:j], g.bossBullets[j+1:]...)
		// 		// 移除玩家子弹
		// 		g.playerBullets = append(g.playerBullets[:i], g.playerBullets[i+1:]...)
		// 		i--
		// 		break
		// 	}
		// }

		// // 检测玩家子弹与敌方子弹的碰撞
		// for j := 0; j < len(g.enemyBullets); j++ {
		// 	if i < 0 {
		// 		break
		// 	}
		// 	if checkCollision(g.playerBullets[i].x, g.playerBullets[i].y, 5, 5, g.enemyBullets[j].x, g.enemyBullets[j].y, 5, 5) {
		// 		// 移除敌方子弹
		// 		g.enemyBullets = append(g.enemyBullets[:j], g.enemyBullets[j+1:]...)
		// 		// 移除玩家子弹
		// 		g.playerBullets = append(g.playerBullets[:i], g.playerBullets[i+1:]...)
		// 		i--
		// 		break
		// 	}
		// }

		// 移除超出屏幕的子弹
		if i >= 0 && i < len(g.playerBullets) {
			if g.playerBullets[i].x < 0 || g.playerBullets[i].x > screenWidth || g.playerBullets[i].y < statusBarHeight || g.playerBullets[i].y > screenHeight {
				g.playerBullets = append(g.playerBullets[:i], g.playerBullets[i+1:]...)
				i--
			}
		}
	}
	return nil
}

// isPlayerTankFollowed 判断玩家坦克是否在尾随Boss坦克
func (g *Game) isPlayerTankFollowed() bool {
	if g.playerTank == nil || g.bossTank == nil {
		return false
	}

	isFollowing := false
	switch g.playerTank.direction {
	case 0: // 上
		if g.playerTank.x == g.bossTank.x && g.playerTank.y > g.bossTank.y {
			isFollowing = true
		}
	case 1: // 右
		if g.playerTank.x < g.bossTank.x && g.playerTank.y == g.bossTank.y {
			isFollowing = true
		}
	case 2: // 下
		if g.playerTank.x == g.bossTank.x && g.playerTank.y < g.bossTank.y {
			isFollowing = true
		}
	case 3: // 左
		if g.playerTank.x > g.bossTank.x && g.playerTank.y == g.bossTank.y {
			isFollowing = true
		}
	}
	return isFollowing
}

func (g *Game) updateBossTank() error {
	if g.bossTank != nil {
		// 检查玩家坦克是否在尾随
		if g.isPlayerTankFollowed() {
			if time.Since(g.lastFollowTime) > bossToleranceTime*time.Second {
				// 玩家坦克尾随超过3秒，Boss坦克转向并射击
				g.bossTank.direction = (g.bossTank.direction + 2) % 4 // 转向180度
				g.bossTankFire()
				g.lastFollowTime = time.Now()
			}
		} else {
			g.lastFollowTime = time.Now()
		}

		// 简单的随机移动逻辑
		if int(ebiten.ActualTPS())%changeDirInterval == 0 && !g.bossTank.directionChanged {
			g.bossTank.direction = rand.Intn(4)
			g.bossTankFire()
			g.bossTank.directionChanged = true
		} else if int(ebiten.ActualTPS())%changeDirInterval != 0 {
			g.bossTank.directionChanged = false
		}

		var newX, newY = g.bossTank.x, g.bossTank.y

		switch g.bossTank.direction {
		case 0:
			if g.bossTank.y > statusBarHeight {
				newY -= tankSpeed
			} else {
				g.bossTank.direction = rand.Intn(4)
				g.bossTankFire()
			}
		case 1:
			if g.bossTank.x < screenWidth-20 {
				newX += tankSpeed
			} else {
				g.bossTank.direction = rand.Intn(4)
				g.bossTankFire()
			}
		case 2:
			if g.bossTank.y < screenHeight-20 {
				newY += tankSpeed
			} else {
				g.bossTank.direction = rand.Intn(4)
				g.bossTankFire()
			}
		case 3:
			if g.bossTank.x > 0 {
				newX -= tankSpeed
			} else {
				g.bossTank.direction = rand.Intn(4)
				g.bossTankFire()
			}
		}

		// 检测与玩家坦克的碰撞
		collision := false
		if g.playerTank != nil {
			if checkCollision(newX, newY, 20, 20, g.playerTank.x, g.playerTank.y, 20, 20) {
				collision = true
			}
		}

		// 检测与墙的碰撞
		for _, wall := range g.walls {
			if checkCollision(newX, newY, 20, 20, wall.x, wall.y, wall.width, wall.height) {
				collision = true
				// 随机改变行进方向
				g.bossTank.direction = rand.Intn(4)
				g.bossTankFire()
				break
			}
		}

		// 检测与敌方坦克的碰撞
		for _, enemyTank := range g.enemyTanks {
			if checkCollision(newX, newY, 20, 20, enemyTank.x, enemyTank.y, 20, 20) {
				collision = true
				// 随机改变行进方向
				g.bossTank.direction = rand.Intn(4)
				g.bossTankFire()
				break
			}
		}

		// 如果没有碰撞，更新敌人坦克位置
		if !collision {
			g.bossTank.x = newX
			g.bossTank.y = newY
		}

		// 简单的随机射击逻辑
		if int(ebiten.ActualTPS())%shootInterval == 0 && !g.bossTank.hasShot {
			g.bossTankFire()
			g.bossTank.hasShot = true
		} else if int(ebiten.ActualTPS())%shootInterval != 0 {
			g.bossTank.hasShot = false
		}
	}
	return nil
}

func (g *Game) bossTankFire() {
	bullet := Bullet{
		x:         g.bossTank.x + 8,
		y:         g.bossTank.y + 8,
		direction: g.bossTank.direction,
	}
	g.bossBullets = append(g.bossBullets, bullet)
}

func (g *Game) updateBossBullets() error {
	// 更新子弹位置
	for i := 0; i < len(g.bossBullets); i++ {
		switch g.bossBullets[i].direction {
		case 0:
			g.bossBullets[i].y -= bulletSpeed
		case 1:
			g.bossBullets[i].x += bulletSpeed
		case 2:
			g.bossBullets[i].y += bulletSpeed
		case 3:
			g.bossBullets[i].x -= bulletSpeed
		}

		// 检测Boss子弹与玩家坦克的碰撞
		if g.playerTank != nil && len(g.bossBullets) > 0 {
			if checkCollision(g.bossBullets[i].x, g.bossBullets[i].y, 5, 5, g.playerTank.x, g.playerTank.y, 20, 20) {
				g.playerTank.health--
				if g.playerTank.health <= 0 {
					// 移除玩家坦克
					g.playerTank = nil
				}
				// 移除子弹
				g.bossBullets = append(g.bossBullets[:i], g.bossBullets[i+1:]...)
				i--
				continue
			}
		}

		// 检测Boss子弹与墙的碰撞
		for j := 0; j < len(g.walls); j++ {
			if i < 0 {
				break
			}
			if checkCollision(g.bossBullets[i].x, g.bossBullets[i].y, 5, 5, g.walls[j].x, g.walls[j].y, g.walls[j].width, g.walls[j].height) {
				g.walls[j].health--
				if g.walls[j].health <= 0 {
					// 移除墙
					g.walls = append(g.walls[:j], g.walls[j+1:]...)
				}
				// 移除子弹
				g.bossBullets = append(g.bossBullets[:i], g.bossBullets[i+1:]...)
				i--
				break
			}
		}

		// 移除超出屏幕的子弹
		if i >= 0 && i < len(g.bossBullets) {
			if g.bossBullets[i].x < 0 || g.bossBullets[i].x > screenWidth || g.bossBullets[i].y < statusBarHeight || g.bossBullets[i].y > screenHeight {
				g.bossBullets = append(g.bossBullets[:i], g.bossBullets[i+1:]...)
				i--
			}
		}
	}
	return nil
}

func (g *Game) updateEnemyTanks() error {
	// 更新敌人坦克状态
	for i := 0; i < len(g.enemyTanks); i++ {
		// 简单的随机移动逻辑
		if int(ebiten.ActualTPS())%changeDirInterval == 0 && !g.enemyTanks[i].directionChanged {
			g.enemyTanks[i].direction = rand.Intn(4)
			g.enemyTankFire(i)
			g.enemyTanks[i].directionChanged = true
		} else if int(ebiten.ActualTPS())%changeDirInterval != 0 {
			g.enemyTanks[i].directionChanged = false
		}

		var newX, newY = g.enemyTanks[i].x, g.enemyTanks[i].y

		switch g.enemyTanks[i].direction {
		case 0:
			if g.enemyTanks[i].y > statusBarHeight {
				newY -= tankSpeed
			} else {
				g.enemyTanks[i].direction = rand.Intn(4)
				g.enemyTankFire(i)
			}
		case 1:
			if g.enemyTanks[i].x < screenWidth-20 {
				newX += tankSpeed
			} else {
				g.enemyTanks[i].direction = rand.Intn(4)
				g.enemyTankFire(i)
			}
		case 2:
			if g.enemyTanks[i].y < screenHeight-20 {
				newY += tankSpeed
			} else {
				g.enemyTanks[i].direction = rand.Intn(4)
				g.enemyTankFire(i)
			}
		case 3:
			if g.enemyTanks[i].x > 0 {
				newX -= tankSpeed
			} else {
				g.enemyTanks[i].direction = rand.Intn(4)
				g.enemyTankFire(i)
			}
		}

		// 检测与玩家坦克的碰撞
		collision := false
		if g.playerTank != nil {
			if checkCollision(newX, newY, 20, 20, g.playerTank.x, g.playerTank.y, 20, 20) {
				collision = true
			}
		}

		// 检测与墙的碰撞
		for _, wall := range g.walls {
			if checkCollision(newX, newY, 20, 20, wall.x, wall.y, wall.width, wall.height) {
				collision = true
				// 随机改变行进方向
				g.enemyTanks[i].direction = rand.Intn(4)
				g.enemyTankFire(i)
				break
			}
		}

		// 检测与Boss坦克的碰撞
		if g.bossTank != nil {
			if checkCollision(newX, newY, 20, 20, g.bossTank.x, g.bossTank.y, 20, 20) {
				collision = true
				// 随机改变行进方向
				g.enemyTanks[i].direction = rand.Intn(4)
				g.enemyTankFire(i)
			}
		}

		// 如果没有碰撞，更新敌人坦克位置
		if !collision {
			g.enemyTanks[i].x = newX
			g.enemyTanks[i].y = newY
		}

		// 简单的随机射击逻辑
		if int(ebiten.ActualTPS())%shootInterval == 0 && !g.enemyTanks[i].hasShot {
			g.enemyTankFire(i)
			g.enemyTanks[i].hasShot = true
		} else if int(ebiten.ActualTPS())%shootInterval != 0 {
			g.enemyTanks[i].hasShot = false
		}
	}
	return nil
}

func (g *Game) enemyTankFire(i int) {
	bullet := Bullet{
		x:         g.enemyTanks[i].x + 8,
		y:         g.enemyTanks[i].y + 8,
		direction: g.enemyTanks[i].direction,
	}
	g.enemyBullets = append(g.enemyBullets, bullet)
}

func (g *Game) updateEnemyBullets() error {
	// 更新子弹位置
	for i := 0; i < len(g.enemyBullets); i++ {
		switch g.enemyBullets[i].direction {
		case 0:
			g.enemyBullets[i].y -= bulletSpeed
		case 1:
			g.enemyBullets[i].x += bulletSpeed
		case 2:
			g.enemyBullets[i].y += bulletSpeed
		case 3:
			g.enemyBullets[i].x -= bulletSpeed
		}

		// 检测敌方子弹与玩家坦克的碰撞
		if g.playerTank != nil && len(g.enemyBullets) > 0 {
			if checkCollision(g.enemyBullets[i].x, g.enemyBullets[i].y, 5, 5, g.playerTank.x, g.playerTank.y, 20, 20) {
				g.playerTank.health--
				if g.playerTank.health <= 0 {
					// 移除玩家坦克
					g.playerTank = nil
				}
				// 移除子弹
				g.enemyBullets = append(g.enemyBullets[:i], g.enemyBullets[i+1:]...)
				i--
				continue
			}
		}

		// 检测子弹与墙的碰撞
		for j := 0; j < len(g.walls); j++ {
			if i < 0 {
				break
			}
			if checkCollision(g.enemyBullets[i].x, g.enemyBullets[i].y, 5, 5, g.walls[j].x, g.walls[j].y, g.walls[j].width, g.walls[j].height) {
				g.walls[j].health--
				if g.walls[j].health <= 0 {
					// 移除墙
					g.walls = append(g.walls[:j], g.walls[j+1:]...)
				}
				// 移除子弹
				g.enemyBullets = append(g.enemyBullets[:i], g.enemyBullets[i+1:]...)
				i--
				break
			}
		}

		// 移除超出屏幕的子弹
		if i >= 0 && i < len(g.enemyBullets) {
			if g.enemyBullets[i].x < 0 || g.enemyBullets[i].x > screenWidth || g.enemyBullets[i].y < statusBarHeight || g.enemyBullets[i].y > screenHeight {
				g.enemyBullets = append(g.enemyBullets[:i], g.enemyBullets[i+1:]...)
				i--
			}
		}
	}
	return nil
}

// Update 更新游戏状态
func (g *Game) Update() error {
	if g.gameOver || g.gameSucc {
		return nil
	}

	g.updatePlayerTank()
	g.updatePlayerBullets()
	g.updateBossTank()
	g.updateBossBullets()
	g.updateEnemyTanks()
	g.updateEnemyBullets()

	// 检测玩家坦克是否被消灭
	if g.playerTank == nil {
		g.gameOver = true
	}

	// 检测Boss坦克是否被消灭
	if g.bossTank == nil {
		g.gameSucc = true
	}

	return nil
}

func (g *Game) drawPlayerTank(screen *ebiten.Image) {
	// 绘制玩家坦克
	if g.playerTank != nil {
		vector.DrawFilledRect(screen, g.playerTank.x, g.playerTank.y, 20, 20, color.RGBA{0, 255, 0, 255}, false)
		switch g.playerTank.direction {
		case 0:
			vector.StrokeLine(screen, g.playerTank.x+10, g.playerTank.y, g.playerTank.x+10, g.playerTank.y-10, 1, color.RGBA{0, 255, 0, 255}, false)
		case 1:
			vector.StrokeLine(screen, g.playerTank.x+20, g.playerTank.y+10, g.playerTank.x+30, g.playerTank.y+10, 1, color.RGBA{0, 255, 0, 255}, false)
		case 2:
			vector.StrokeLine(screen, g.playerTank.x+10, g.playerTank.y+20, g.playerTank.x+10, g.playerTank.y+30, 1, color.RGBA{0, 255, 0, 255}, false)
		case 3:
			vector.StrokeLine(screen, g.playerTank.x, g.playerTank.y+10, g.playerTank.x-10, g.playerTank.y+10, 1, color.RGBA{0, 255, 0, 255}, false)
		}
	}
}

func (g *Game) drawPlayerBullets(screen *ebiten.Image) {
	// 绘制玩家子弹
	for _, bullet := range g.playerBullets {
		vector.DrawFilledRect(screen, bullet.x, bullet.y, 5, 5, color.RGBA{0, 255, 0, 255}, false)
	}
}

func (g *Game) drawBossTank(screen *ebiten.Image) {
	if g.bossTank != nil {
		vector.DrawFilledRect(screen, g.bossTank.x, g.bossTank.y, 20, 20, color.RGBA{255, 0, 0, 255}, false)
		switch g.bossTank.direction {
		case 0:
			vector.StrokeLine(screen, g.bossTank.x+10, g.bossTank.y, g.bossTank.x+10, g.bossTank.y-10, 1, color.RGBA{255, 0, 0, 255}, false)
		case 1:
			vector.StrokeLine(screen, g.bossTank.x+20, g.bossTank.y+10, g.bossTank.x+30, g.bossTank.y+10, 1, color.RGBA{255, 0, 0, 255}, false)
		case 2:
			vector.StrokeLine(screen, g.bossTank.x+10, g.bossTank.y+20, g.bossTank.x+10, g.bossTank.y+30, 1, color.RGBA{255, 0, 0, 255}, false)
		case 3:
			vector.StrokeLine(screen, g.bossTank.x, g.bossTank.y+10, g.bossTank.x-10, g.bossTank.y+10, 1, color.RGBA{255, 0, 0, 255}, false)
		}
	}
}

func (g *Game) drawBossBullets(screen *ebiten.Image) {
	// 绘制敌人子弹
	for _, bullet := range g.bossBullets {
		vector.DrawFilledRect(screen, bullet.x, bullet.y, 5, 5, color.RGBA{255, 0, 0, 255}, false)
	}
}

func (g *Game) drawEnemyTanks(screen *ebiten.Image) {
	// 绘制敌人坦克
	for _, enemyTank := range g.enemyTanks {
		vector.DrawFilledRect(screen, enemyTank.x, enemyTank.y, 20, 20, color.RGBA{255, 182, 193, 255}, false)
		switch enemyTank.direction {
		case 0:
			vector.StrokeLine(screen, enemyTank.x+10, enemyTank.y, enemyTank.x+10, enemyTank.y-10, 1, color.RGBA{255, 182, 193, 255}, false)
		case 1:
			vector.StrokeLine(screen, enemyTank.x+20, enemyTank.y+10, enemyTank.x+30, enemyTank.y+10, 1, color.RGBA{255, 182, 193, 255}, false)
		case 2:
			vector.StrokeLine(screen, enemyTank.x+10, enemyTank.y+20, enemyTank.x+10, enemyTank.y+30, 1, color.RGBA{255, 182, 193, 255}, false)
		case 3:
			vector.StrokeLine(screen, enemyTank.x, enemyTank.y+10, enemyTank.x-10, enemyTank.y+10, 1, color.RGBA{255, 182, 193, 255}, false)
		}
	}
}

func (g *Game) drawEnemyBullets(screen *ebiten.Image) {
	// 绘制敌人子弹
	for _, bullet := range g.enemyBullets {
		vector.DrawFilledRect(screen, bullet.x, bullet.y, 5, 5, color.RGBA{255, 182, 193, 255}, false)
	}
}

func (g *Game) drawWalls(screen *ebiten.Image) {
	// 绘制墙壁
	for _, wall := range g.walls {
		vector.DrawFilledRect(screen, wall.x, wall.y, wall.width, wall.height, color.RGBA{128, 128, 128, 255}, false)
	}
}

func (g *Game) drawStatusBar(screen *ebiten.Image) {
	// 绘制状态栏
	vector.DrawFilledRect(screen, 0, 0, screenWidth, 20, color.RGBA{192, 192, 192, 255}, false)

	const (
		fontSize = 14
	)

	face := &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   fontSize,
	}
	op := &text.DrawOptions{}
	op.ColorScale.ScaleWithColor(color.RGBA{0, 0, 255, 255})

	var msg string
	// // 绘制敌方坦克总数
	// msg = fmt.Sprintf("敌方坦克总数: %d", g.enemyTankCount)
	// op.GeoM.Translate(2, 1)
	// text.Draw(screen, msg, face, op)

	// // 绘制敌方出动坦克数
	// msg = fmt.Sprintf("敌方出动坦克数: %d", len(g.enemyTanks))
	// op.GeoM.Translate(120, 1)
	// text.Draw(screen, msg, face, op)

	// 绘制玩家坦克生命值
	if g.playerTank != nil {
		msg = fmt.Sprintf("玩家生命值: %d", g.playerTank.health)
		op.GeoM.Translate(2, 1)
		text.Draw(screen, msg, face, op)
	}

	// 绘制Boss坦克生命值
	if g.bossTank != nil {
		msg = fmt.Sprintf("敌方生命值: %d", g.bossTank.health)
		op.GeoM.Translate(120, 1)
		text.Draw(screen, msg, face, op)
	}
}

// Draw 绘制游戏画面
func (g *Game) Draw(screen *ebiten.Image) {
	if g.gameOver {
		ebitenutil.DebugPrint(screen, "GAME OVER!")
		return
	}

	if g.gameSucc {
		ebitenutil.DebugPrint(screen, "YOU WIN!")
		return
	}

	g.drawStatusBar(screen)
	g.drawPlayerTank(screen)
	g.drawPlayerBullets(screen)
	g.drawBossTank(screen)
	g.drawBossBullets(screen)
	g.drawEnemyTanks(screen)
	g.drawEnemyBullets(screen)
	g.drawWalls(screen)
}

// Layout 返回游戏画面的布局
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
