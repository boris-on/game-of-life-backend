package game

import (
	"encoding/json"
	"fmt"
	"image/color"
)

type World struct {
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Tps       int     `json:"tps"`
	TpsCount  int     `json:"tps_count"`
	Area      [][]int `json:"area"`
	Units     `json:"units"`
	Colors    map[color.RGBA]int `json:"colors"`
	UnitCount int                `json:"unit_count"`
	MyID      int                `json:"myid"`
	IsServer  bool               `json:"iserver"`
	// colors := []color.Color{
	// 	color.RGBA{0, 0, 255, 0},
	// 	color.RGBA{0, 255, 0, 0},
	// 	color.RGBA{0, 255, 255, 0},
	// 	color.RGBA{255, 0, 0, 0},
	// 	color.RGBA{255, 0, 255, 0},
	// 	color.RGBA{255, 255, 0, 0},
	// }
}

type Units map[int]*Unit

type Unit struct {
	ID    int        `json:"id"`
	Color color.RGBA `json:"color"`
}

type Event struct {
	Type string `json:"type"`
	Data interface{}
}

type EventConnect struct {
	Unit `json:"unit"`
}

type EventFillCell struct {
	ID int `json:"id"`
	X  int `json:"x"`
	Y  int `json:"y"`
}

type EventInit struct {
	ID    int     `json:"id"`
	Units Units   `json:"units"`
	Area  [][]int `json:"area"`
}

type EventConfirmUpdate struct {
	Area [][]int `json:"area"`
}

type EventDisconnect struct {
	ID int `json:"id"`
}

const EventTypeConnect = "connect"
const EventTypeFillCell = "fill"
const EventTypeInit = "init"
const EventTypeUpdateCells = "update"
const EventTypeConfirmUpdate = "confirm"
const EventTypeDisconnect = "disconnect"

func nTrue(val ...int) (int, int) {
	sum := 0
	id := 0
	idMap := make(map[int]int, 8)
	for i := range val {
		if val[i] != 0 {
			sum++
			idMap[val[i]]++
		}
	}

	for i := range idMap {
		id = i
		break
	}

	for i := range idMap {
		if idMap[i] == 2 || idMap[i] == 3 {
			id = i
		}
	}

	return sum, id
}

func (world *World) UpdateCells() {
	buffer := make([][]int, world.Height)
	for i := range buffer {
		buffer[i] = make([]int, world.Width)
	}
	for y := 1; y < world.Height-1; y++ {
		for x := 1; x < world.Width-1; x++ {
			// buffer[y][x] = false
			sum, id := nTrue(world.Area[y][x-1], world.Area[y-1][x-1], world.Area[y-1][x],
				world.Area[y-1][x+1], world.Area[y][x+1], world.Area[y+1][x+1], world.Area[y+1][x], world.Area[y+1][x-1])
			switch {
			case sum < 2:
				buffer[y][x] = 0
			case (sum == 2 || sum == 3) && world.Area[y][x] != 0:
				buffer[y][x] = id
			case sum > 3:
				buffer[y][x] = 0
			case sum == 3:
				buffer[y][x] = id
			}
		}
	}

	temp := buffer
	// buffer = g.world.area
	world.Area = temp
}

func (world *World) HandleEvent(event *Event) {

	switch event.Type {
	case EventTypeInit:
		str, _ := json.Marshal(event.Data)
		var ev EventInit
		json.Unmarshal(str, &ev)

		world.MyID = ev.ID
		world.Units = ev.Units

		for i := range ev.Area {
			for j := range ev.Area[i] {
				world.Area[i][j] = ev.Area[i][j]
			}
		}

	case EventTypeConnect:
		str, _ := json.Marshal(event.Data)
		var ev EventConnect
		json.Unmarshal(str, &ev)

		world.Units[ev.ID] = &ev.Unit

	case EventTypeDisconnect:
		str, _ := json.Marshal(event.Data)
		var ev EventDisconnect
		json.Unmarshal(str, &ev)
		fmt.Println("before deleting", world.Units, ev.ID)
		delete(world.Units, ev.ID)
		fmt.Println("after deleting", world.Units)
		for clr := range world.Colors {
			if world.Colors[clr] == ev.ID {
				world.Colors[clr] = 0
				break
			}
		}

		for i := range world.Area {
			for j := range world.Area[i] {
				if world.Area[i][j] == ev.ID {
					world.Area[i][j] = 0
				}
			}
		}
	case EventTypeFillCell:
		str, _ := json.Marshal(event.Data)
		var ev EventFillCell
		json.Unmarshal(str, &ev)
		world.Area[ev.Y][ev.X] = ev.ID
		world.Area[ev.Y][ev.X-1] = ev.ID
		world.Area[ev.Y+1][ev.X] = ev.ID
		world.Area[ev.Y-1][ev.X] = ev.ID
		world.Area[ev.Y-1][ev.X+1] = ev.ID
	}
}

func (world *World) AddUnit() *Unit {

	unitId := world.UnitCount + 1
	var unitColor color.RGBA
	world.UnitCount++
	for clr := range world.Colors {
		if world.Colors[clr] == 0 {
			unitColor = clr
			world.Colors[clr] = unitId
			break
		}
	}

	unit := &Unit{
		ID:    unitId,
		Color: unitColor,
	}

	world.Units[unitId] = unit
	return unit

}
