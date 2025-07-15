package main

import "strings"

// Вывести пользователю, какие поля свободны в текущей игре
func renderBoard(game *Game, message string) OutgoingMessage {
	var sb strings.Builder
	for _, row := range game.Board {
		for _, cell := range row {
			if cell == "" {
				sb.WriteString(".")
			} else {
				sb.WriteString(cell)
			}
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}

	coords := []string{"A1", "A2", "A3", "B1", "B2", "B3", "C1", "C2", "C3"}
	var buttons []Button
	for _, coord := range coords {
		x, y := parseCoord(coord)
		if game.Board[x][y] == "" && !game.Finished {
			buttons = append(buttons, Button{
				Text:   coord,
				Action: "Move:" + coord,
			})
		}
	}

	return OutgoingMessage{
		UserID:  game.Turn,
		Text:    message + "\n" + sb.String(),
		Buttons: buttons,
	}
}

// Перевести координаты из строки в массив
func parseCoord(coord string) (int, int) {
	m := map[string][2]int{
		"A1": {0, 0}, "A2": {0, 1}, "A3": {0, 2},
		"B1": {1, 0}, "B2": {1, 1}, "B3": {1, 2},
		"C1": {2, 0}, "C2": {2, 1}, "C3": {2, 2},
	}
	if val, ok := m[coord]; ok {
		return val[0], val[1]
	}
	return -1, -1
}
