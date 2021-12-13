package main

import (
	"encoding/json"
	"fmt"
	"log"
	rand2 "math/rand"
	"net/http"
	"os"
)

func main() {
	port := "8080"
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}
	http.HandleFunc("/", handler)

	log.Printf("starting server on port :%s", port)
	err := http.ListenAndServe(":"+port, nil)
	log.Fatalf("http listen error: %v", err)
}

func handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		fmt.Fprint(w, "Let the battle begin!")
		return
	}

	var v ArenaUpdate
	defer req.Body.Close()
	d := json.NewDecoder(req.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&v); err != nil {
		log.Printf("WARN: failed to decode ArenaUpdate in response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := play(v)
	fmt.Fprint(w, resp)
}

func play(input ArenaUpdate) (response string) {
	log.Printf("IN: %#v", input)
	escape := shouldEscape(input)

	if !escape && shouldThrow(input.Links.Self.Href, input) {
		return "T"
	}
	occupied := isTileOccupied(input.Links.Self.Href, input)
	log.Printf("Es: %v.  Occ: %v", escape, occupied)
	if !occupied && (escape || shouldGo(input.Links.Self.Href, input)) {
		return "F"
	}
	if shouldRight(input.Links.Self.Href, input) {
		return "R"
	}
	if shouldLeft(input.Links.Self.Href, input) {
		return "L"
	}
	return move(input, occupied)
}

func shouldEscape(input ArenaUpdate) bool {
	myself := input.Arena.State[input.Links.Self.Href]
	c := 0
	if myself.WasHit {
		for player, state := range input.Arena.State {
			if player != input.Links.Self.Href {
				if inRange(state, myself, state.Direction, 3) {
					c++
				}
			}
		}
	}
	return c > 0
}

func move(input ArenaUpdate, occupied bool) string {

	myself := input.Arena.State[input.Links.Self.Href]
	return chooseMove(
		myself.X, myself.Y,
		input.Arena.Dimensions[0],
		input.Arena.Dimensions[1],
		myself.Direction,
		occupied,
	)
}

func chooseMove(x int, y int, max_x int, max_y int, direction string, occupied bool) string {

	if x == 0 || x == (max_x-1) {
		if y == 0 {
			if x == 0 {
				if direction == "N" {
					return "R"
				} else if direction == "W" {
					return "L"
				} else if direction == "S" {
					return choose([]string{"L", "F"}, occupied)
				} else {
					return choose([]string{"R", "F"}, occupied)
				}
			} else {
				if direction == "N" {
					return "L"
				} else if direction == "E" {
					return "R"
				} else if direction == "S" {
					return choose([]string{"R", "F"}, occupied)
				} else {
					return choose([]string{"L", "F"}, occupied)
				}
			}
		}
		if y == max_y-1 {
			if x == 0 {
				if direction == "S" {
					return "L"
				} else if direction == "W" {
					return "R"
				} else if direction == "N" {
					return choose([]string{"R", "F"}, occupied)
				} else {
					return choose([]string{"L", "F"}, occupied)
				}
			} else {
				if direction == "S" {
					return "R"
				} else if direction == "E" {
					return "L"
				} else if direction == "N" {
					return choose([]string{"L", "F"}, occupied)
				} else {
					return choose([]string{"R", "F"}, occupied)
				}
			}
		}
		if y == (max_y - 1) {
			if direction == "N" || direction == "W" {
				return "L"
			} else {
				return "R"
			}
		}
		if x == 0 {
			if direction == "N" {
				return []string{"R", "F"}[rand2.Intn(2)]
			} else if direction == "S" {
				return choose([]string{"L", "F"}, occupied)
			}
			if direction == "W" {
				return choose([]string{"L", "R"}, false)
			}
		}
		if x == max_x-1 {
			if direction == "N" {
				return choose([]string{"L", "F"}, occupied)
			} else if direction == "S" {
				return choose([]string{"R", "F"}, occupied)
			}
			if direction == "E" {
				return choose([]string{"L", "R"}, false)
			}
		}
	}

	if y == 0 {
		if direction == "E" {
			return choose([]string{"L", "F"}, occupied)
		} else if direction == "W" {
			return choose([]string{"R", "F"}, occupied)
		}
		if direction == "N" {
			return choose([]string{"L", "R"}, false)
		}
	}
	if y == max_y-1 {
		if direction == "E" {
			return []string{"R", "F"}[rand2.Intn(2)]
		} else if direction == "W" {
			return []string{"L", "F"}[rand2.Intn(2)]
		}
		if direction == "S" {
			return choose([]string{"L", "R"}, false)
		}
	}
	return choose([]string{"L", "R", "F"}, occupied)
}

func choose(options []string, occupied bool) string {
	l := len(options)
	if occupied {
		l -= 1
	}
	return options[rand2.Intn(l)]
}

func shouldRight(me string, input ArenaUpdate) bool {
	myself := input.Arena.State[me]
	direction := map[string]string{
		"S": "W",
		"W": "N",
		"N": "E",
		"E": "S",
	}
	for player, state := range input.Arena.State {
		if player != me {
			if inRange(myself, state, direction[myself.Direction], 3) {
				return true
			}
		}
	}
	return false
}

func shouldLeft(me string, input ArenaUpdate) bool {
	myself := input.Arena.State[me]
	direction := map[string]string{
		"W": "S",
		"N": "W",
		"E": "N",
		"S": "E",
	}
	for player, state := range input.Arena.State {
		if player != me {
			if inRange(myself, state, direction[myself.Direction], 3) {
				return true
			}
		}
	}
	log.Printf("should lefg %s", direction[myself.Direction])
	return false
}

func shouldGo(me string, input ArenaUpdate) bool {
	for player, state := range input.Arena.State {
		if player != me {
			if inRange(input.Arena.State[me], state, input.Arena.State[me].Direction, 5) {
				return true
			}
		}
	}
	return false
}

func isTileOccupied(me string, input ArenaUpdate) bool {
	myself := input.Arena.State[me]
	newX := myself.X
	newY := myself.Y
	switch myself.Direction {
	case "S":
		newY += 1
	case "N":
		newY -= 1
	case "E":
		newX += 1
	case "W":
		newX -= 1
	}	
	if newX >= input.Arena.Dimensions[0] || newX < 0 || newY >= input.Arena.Dimensions[1] || newY < 0 {
		return true
	}
	for player, state := range input.Arena.State {
		if player != me {
			if state.X == newX && state.Y == newY {
				return true
			}
		}
	}

	return false
}
func shouldThrow(me string, input ArenaUpdate) bool {
	if input.Arena.State[me].WasHit {
		if 50 < rand2.Intn(100) {
			return false
		}
	}
	for player, state := range input.Arena.State {
		if player != me {
			if inRange(input.Arena.State[me], state, input.Arena.State[me].Direction, 3) {
				return true
			}
		}
		log.Printf("R: %#v", player)
	}
	return false
}

func inRange(me PlayerState, other PlayerState, direction string, dis int) bool {
	r := -1
	if direction == "N" && me.X == other.X {
		r = me.Y - other.Y
	}
	if direction == "S" && me.X == other.X {
		r = other.Y - me.Y
	}
	if direction == "E" && me.Y == other.Y {
		r = other.X - me.X
	}
	if direction == "W" && me.Y == other.Y {
		r = me.X - other.X
	}
	log.Printf("R: %s %#v", direction, r)
	return r > 0 && r <= dis
}
