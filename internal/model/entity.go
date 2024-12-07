package model

import (
	l "artificialLifeGo/internal/logger"
	"fmt"
	"strconv"
)

// NewEntity возвращает живую сущность(Entity) с координатами x, y.
func NewEntity(ID, x, y, longDNA int) *Entity {
	return &Entity{
		ID,
		0,
		100,
		true,
		0,
		Coordinates{
			x,
			y,
		},
		*NewDNA(longDNA),
		newBrain(),
	}
}

// Run отвечает за исполнение генетического кода в DNA.Array.
// Возвращает nil или критическую ошибку
func (e *Entity) Run(w *World) {
	l.Ent.Debug("id " + strconv.Itoa(e.ID) + " is run his genocode")
	//если бот мёрт, вылетаем с ошибкой
	if !e.Live {
		l.Ent.Debug("ID:" + strconv.Itoa(e.ID) + "cant run - dead")
		return
	}

	//уменьшаем энергию бота перед выполнение генокода
	// "Деньги в перёд"
	e.Energy--
	e.Age++

	err := e.run(e, w)
	if err != nil {
		l.Ent.Error("id" + strconv.Itoa(e.ID) + " " + err.Error())
		return
	}

	//Берём клетку, где находиться сущность
	cell, err := w.GetCellData(e.Coordinates)
	if err != nil {
		l.Ent.Error(strconv.Itoa(e.ID) + " " + err.Error())
		return
	}

	//Проверяем колличество яда, много - умираем
	if cell.Poison >= pLevelDed {
		e.die(w)
		l.Ent.Info("ID:" + strconv.Itoa(e.ID) + " die inside poison")
		return
	}

	//Если энергии не осталось - умираем
	if e.Energy <= 0 {
		e.die(w)
		l.Ent.Info("ID:" + strconv.Itoa(e.ID) + " die without energy")
		return
	}
}

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

// move отвечает за передвижение сущности(Entity) из одной клетки(Cell) мира(World) в другую.
// Возвращает nil или ошибку.
func (e *Entity) move(w *World) error {
	//получаем координаты, куда хотим переместиться
	newCord := Sum(
		viewCell(e.turn),
		e.Coordinates,
	)
	//перемещаемся в новые координаты
	if err := w.MoveEntity(e.Coordinates, newCord, e); err != nil {
		return err
	}
	e.Coordinates = newCord
	return nil
}

// look отвечает за получение данных из другой клетки(Cell). Возвращает номер
// сдвига Entity.DNA.Pointer или ошибку.
func (e *Entity) look(w *World) (int, error) {
	//константы ответов на что мы смотрим
	const (
		isError = iota
		isEmpty
		isFood
		isWall
		isEntity
	)

	//получаем координаты, куда хотим посмотреть
	newCord := Sum(
		viewCell(e.turn),
		e.Coordinates)
	//смотрим что там
	cell, err := w.GetCellData(newCord)
	if err != nil {
		return isError, err
	}

	//Определяем тип возврата
	switch cell.Types {
	case EmptyCell:
		if cell.Entity != nil {
			return isEntity, nil
		} else {
			return isEmpty, nil
		}
	case FoodCell:
		return isFood, nil
	case WallCell:
		return isWall, nil
	default:
		return isError, fmt.Errorf("cell type is %v, I dont't know this type", cell.Types)
	}
}

// get отвечает за взаимодействие сущности(Entity) с окружением
// таким как: взять, съесть и тп. Возвращает nil или ошибку.
func (e *Entity) get(w *World) error {
	//получаем координаты для взятия
	newCord := Sum(
		viewCell(e.turn),
		e.Coordinates)
	//смотрим что там
	cell, err := w.GetCellData(newCord)
	if err != nil {
		return err
	}
	//совераем действие в зависимости от типа клетки
	switch cell.Types {
	case EmptyCell:
		if err = e.attack(cell); err != nil {
			return err
		}
	case FoodCell:
		//сначала меняем тип клетки
		cell.Types = EmptyCell
		//а потом увеличиваем энергию
		e.Energy += EnergyPoint
	case WallCell:
		e.Energy -= EnergyPoint
	default:
		return fmt.Errorf("cell type is %v, I dont't know this type", cell.Types)
	}
	return nil
}

// attack отвечает за убийство сущности(Entity) в клетке(Cell) и передачи энергии сущности(Entity),
// вы звавщей функцию. Ничего не возвращает.
func (e *Entity) attack(cell *Cell) error {
	if cell.Entity == nil {
		return fmt.Errorf("attack is fall - not entity")
	}
	energy := cell.Entity.Energy
	cell.Entity.Live = false
	cell.Entity = nil
	e.Energy = energy
	return nil
}

// rotation отвечает за смену угла взгляда на заданное число.
// Повороты зациклены.
func (e *Entity) rotation(turnCount turns) {
	e.turn = (e.turn + turnCount) % 8
}

// recycling отвечает за получение энергии из загрязнения окружающей среды.
// Возвращает nil или ошибку.
func (e *Entity) recycling(w *World) error {

	//получаем координаты переработки
	newCord := viewCell(e.turn)
	//смотрим что там
	cell, err := w.GetCellData(
		Sum(
			newCord,
			e.Coordinates))
	if err != nil {
		return err
	}

	//Расчитываем размер очистки клетки
	var dPoison = 0
	if cell.Poison >= pLevel4 {
		dPoison = EnergyPoint * 2
	} else if cell.Poison >= pLevel3 {
		dPoison = EnergyPoint
	} else if cell.Poison >= pLevel2 {
		dPoison = EnergyPoint / 2
	} else if cell.Poison >= pLevel1 {
		dPoison = EnergyPoint / 5
	}

	//очищаем клетку
	if err = w.SetCellPoison(newCord, cell.Poison-dPoison); err != nil {
		return err
	}

	return nil
}

// reproduction is todo!
func (e *Entity) reproduction() error {
	return nil
}

// jump обеспечивает зацикленный прыжок по DNA.Array.
func (e *Entity) jump() {
	e.Pointer += (e.Pointer + e.Array[e.Pointer]) % LengthDNA
}

// loopPointer обеспечивает зацикленность DNA.Pointer.
func (e *Entity) loopPointer() {
	e.Pointer = e.Pointer % LengthDNA
}

// die устанавливает бота в умершее состояние.
// Удаляет ссылку на бота из мира. Первый уровень защиты от умерших ботов.
func (e *Entity) die(w *World) {
	e.Live = false
	e.Energy = 0
	//очищаем клетку от сущности
	_ = w.SetCellEntity(e.Coordinates, nil)
}
