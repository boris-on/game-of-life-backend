package server

import (
	"image/color"
	"os"

	"github.com/boris-on/game-of-live-backend/game"
	"github.com/gin-gonic/gin"
)

func main() {
	world := &game.World{
		Units: game.Units{},
		Colors: map[color.RGBA]int{
			color.RGBA{0, 0, 255, 0}:   0,
			color.RGBA{0, 255, 0, 0}:   0,
			color.RGBA{0, 255, 255, 0}: 0,
			color.RGBA{255, 255, 0, 0}: 0,
		},
		Width:    250,
		Height:   250,
		Area:     make([][]int, 250),
		IsServer: true,
	}

	for i := range world.Area {
		world.Area[i] = make([]int, 250)
	}

	hub := newHub()
	go hub.run()
	r := gin.New()
	r.GET("/ws", func(hub *Hub, world *game.World) gin.HandlerFunc {
		return gin.HandlerFunc(func(c *gin.Context) {
			serveWs(hub, world, c.Writer, c.Request)
		})
	}(hub, world))
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	}
	r.Run(port)
}