package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/boris-on/game-of-life-backend/game"
	"github.com/hajimehoshi/ebiten"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var world game.World
var frame int

func init() {

	world = game.World{
		Width:  250,
		Height: 250,
		Units:  game.Units{},
		Area:   make([][]int, 250),
	}

	for i := range world.Area {
		world.Area[i] = make([]int, 250)
	}

}

func update(c *websocket.Conn) func(screen *ebiten.Image) error {
	return func(screen *ebiten.Image) error {
		frame++
		screen.Clear()

		for y := range world.Area {
			for x := range world.Area[y] {
				unitId := world.Area[y][x]
				clr := color.RGBA{0, 0, 0, 0}
				if unitId != 0 {
					for i := range world.Units {
						if world.Units[i].ID == unitId {
							clr = world.Units[i].Color
						}
					}
				}

				screen.Set(x, y, clr)

			}
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			ctx := context.Background()
			posX, posY := ebiten.CursorPosition()
			wsjson.Write(ctx, c, game.Event{
				Type: game.EventTypeFillCell,
				Data: game.EventFillCell{
					X:  posX,
					Y:  posY,
					ID: world.MyID,
				},
			})
		}

		return nil
	}
}

func main() {
	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, "https://mighty-earth-39320.herokuapp.com/ws", nil)
	// c, _, err := websocket.Dial(ctx, "ws://127.0.0.1:3000/ws", nil)
	if err != nil {
		fmt.Println(err)
	}
	c.SetReadLimit(524288000)
	go func(c *websocket.Conn) {
		defer func() {
			ctx := context.Background()
			wsjson.Write(ctx, c, game.Event{
				Type: game.EventTypeDisconnect,
				Data: game.EventDisconnect{
					ID: world.MyID,
				},
			})
			c.Close(websocket.StatusNormalClosure, "")

		}()
		go func() {
			for {
				go world.UpdateCells()
				time.Sleep(50 * time.Millisecond)
			}
		}()
		for {
			var event game.Event
			err = wsjson.Read(ctx, c, &event)
			if err != nil {
				fmt.Println(err)
				return
			}
			world.HandleEvent(&event)
			log.Println(event)
		}

	}(c)
	ebiten.SetRunnableInBackground(true)
	ebiten.Run(update(c), 250, 250, 2, "name")
}
