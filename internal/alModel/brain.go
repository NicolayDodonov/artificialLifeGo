package alModel

import (
	l "artificialLifeGo/internal/logger"
	"fmt"
	"strconv"
)

// brain - интерфейс обработчика генокода в DNA
type brain interface {
	run(*Entity, *World) error
}

// newBrain создаёт структуру brain в зависимости от состояния TypeBrain.
func newBrain() brain {
	switch TypeBrain {
	case "16":
		return brain16{}
	case "64":
		return brain64{}
	case "nero":
		return brainNero{}
	default:
		return brain0{}
	}
}

// brain0 структура заглушка
type brain0 struct{}

func (brain0) run(e *Entity, w *World) error {
	_ = e.move(w)
	e.rotation(left)
	return nil
}

// brain16 представляет числа от 0 до 7 как команды
// остальные числа как безусловный переход
type brain16 struct{}

func (brain16) run(e *Entity, w *World) error {
	/*
		выполняем генетический код
		не все команды равноценны по сложности, по этому
		выполняем их со счётчиком frameCount. Это создёт
		более сложное поведение ботов.
	*/
	for frameCount := 0; frameCount < maxFC; {
		//создаём переменную некретической ошибки
		var errGen error

		//считываем генокод по указателю
		switch e.Array[e.Pointer] {
		case move:
			errGen = e.move(w)
			frameCount += middleFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " move")
		case look:
			//функционал логического перехода
			var dPointer int
			dPointer, errGen = e.look(w)
			e.Pointer += dPointer - 1
			frameCount += smallFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " look")
		case get:
			errGen = e.get(w)
			frameCount += middleFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " get")
		case rotatedLeft:
			e.rotation(left)
			frameCount += minFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " turns left")
		case rotatedRight:
			e.rotation(right)
			frameCount += minFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " turns right")
		case recycling:
			errGen = e.recycling(w)
			frameCount += middleFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " recycling")
		case reproduction:
			errGen = e.reproduction()
			frameCount += maxFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " make new bot")
		default:
			e.jump()
			frameCount += minFC
		}
		//Логгируем некретические ошибки генокода
		if errGen != nil {
			l.Ent.Debug("id:" + strconv.Itoa(e.ID) + " " + errGen.Error())
		}

		//увеличиваем программно-генетический счётчик
		e.Pointer++
		e.loopPointer()
	}
	return nil
}

// brain64 представляет первые 32 числа как команды четырёх групп:
// движение (0-7), взгляд (8-15), взять (16-23) и поворот (24-31).
// остальные команды как безусловный переход.
type brain64 struct{}

func (brain64) run(e *Entity, w *World) error {
	if MaxGen < 64 {
		return fmt.Errorf("max gen is small! Max gen: %v", MaxGen)
	}

	for frameCount := 0; frameCount < 10; {
		var errGen error
		command := e.Array[e.Pointer]

		switch command / 8 {
		case move: // 0 to 7
			//поворачиваем
			nowTurn := e.turn
			e.rotation(turns(command % 8))

			errGen = e.move(w)
			e.turn = nowTurn
			frameCount += middleFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " move")
		case look: //8 to 15
			nowTurn := e.turn
			e.rotation(turns(command % 8))

			var dPointer int
			dPointer, errGen = e.look(w)
			e.Pointer += dPointer - 1
			e.turn = nowTurn
			frameCount += smallFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " look")
		case get: //16 to 23
			nowTurn := e.turn
			e.rotation(turns(command % 8))

			errGen = e.get(w)
			e.turn = nowTurn
			frameCount += middleFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " get")
		case rotatedLeft: //24 to 31
			e.rotation(turns(command % 8))
			frameCount += minFC

			l.Ent.Debug("id " + strconv.Itoa(e.ID) + " turns left")
		/*
			case 4: //32 to 39

				nowTurn := e.turn
				e.rotation(turns(command % 8))

				errGen = e.recycling(w)
				e.turn = nowTurn
				frameCount += middleFC

				l.Ent.Debug("id " + strconv.Itoa(e.ID) + " recycling")
		*/
		default: //32 to 64
			e.jump()
			frameCount += minFC
		}

		e.Pointer++
		e.loopPointer()

		if errGen != nil {
			l.Ent.Debug("id:" + strconv.Itoa(e.ID) + " " + errGen.Error())
		}
	}
	return nil
}

// brainNero заглушка на будующее. Интерпритатор на основе нейросети.
type brainNero struct{}

func (brainNero) run(e *Entity, w *World) error {
	return nil
}
