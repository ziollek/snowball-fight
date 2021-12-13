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

	if shouldThrow(input.Links.Self.Href, input) {
		return "T"
	}
	if shouldGo(input.Links.Self.Href, input) {
		return "F"
	}
	if shouldRight(input.Links.Self.Href, input) {
		return "R"
	}
	if shouldLeft(input.Links.Self.Href, input) {
		return "L"
	}
	log.Printf("RANDOM: ")
	return move(input)
}

func move(input ArenaUpdate) string {

	myself := input.Arena.State[input.Links.Self.Href]
	return chooseMove(
		myself.X, myself.Y,
		input.Arena.Dimensions[0],
		input.Arena.Dimensions[1],
		myself.Direction,
	)
}

func chooseMove(x int, y int, max_x int, max_y int, direction string) string {

	if x == 0 || x == (max_x-1) {
		if y == 0 {
			if x == 0 {
				if direction == "N" {
					return "R"
				} else if direction == "W" {
					return "L"
				} else if direction == "S" {
					return []string{"L", "F"}[rand2.Intn(2)]
				} else {
					return []string{"R", "F"}[rand2.Intn(2)]
				}
			} else {
				if direction == "N" {
					return "L"
				} else if direction == "E" {
					return "R"
				} else if direction == "S" {
					return []string{"R", "F"}[rand2.Intn(2)]
				} else {
					return []string{"L", "F"}[rand2.Intn(2)]
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
					return []string{"R", "F"}[rand2.Intn(2)]
				} else {
					return []string{"L", "F"}[rand2.Intn(2)]
				}
			} else {
				if direction == "S" {
					return "R"
				} else if direction == "E" {
					return "L"
				} else if direction == "N" {
					return []string{"L", "F"}[rand2.Intn(2)]
				} else {
					return []string{"R", "F"}[rand2.Intn(2)]
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
				return []string{"L", "F"}[rand2.Intn(2)]
			}
			if direction == "W" {
				return []string{"L", "R"}[rand2.Intn(2)]
			}
		}
		if x == max_x-1 {
			if direction == "N" {
				return []string{"L", "F"}[rand2.Intn(2)]
			} else if direction == "S" {
				return []string{"R", "F"}[rand2.Intn(2)]
			}
			if direction == "E" {
				return []string{"L", "R"}[rand2.Intn(2)]
			}
		}
	}

	if y == 0 {
		if direction == "E" {
			return []string{"L", "F"}[rand2.Intn(2)]
		} else if direction == "W" {
			return []string{"R", "F"}[rand2.Intn(2)]
		}
		if direction == "N" {
			return []string{"L", "R"}[rand2.Intn(2)]
		}
	}
	if y == max_y-1 {
		if direction == "E" {
			return []string{"R", "F"}[rand2.Intn(2)]
		} else if direction == "W" {
			return []string{"L", "F"}[rand2.Intn(2)]
		}
		if direction == "S" {
			return []string{"L", "R"}[rand2.Intn(2)]
		}
	}
	return []string{"L", "R", "F"}[rand2.Intn(3)]
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
	return false
}

func shouldGo(me string, input ArenaUpdate) bool {
	for player, state := range input.Arena.State {
		if player != me {
			if inRange(input.Arena.State[me], state, input.Arena.State[me].Direction, 5) {
				return true
			}
		}
		log.Printf("R: %#v", player)
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
	if me.Direction == "N" && me.X == other.X {
		r = me.Y - other.Y
	}
	if me.Direction == "S" && me.X == other.X {
		r = other.Y - me.Y
	}
	if me.Direction == "E" && me.Y == other.Y {
		r = other.X - me.X
	}
	if me.Direction == "W" && me.Y == other.Y {
		r = me.X - other.X
	}
	log.Printf("R: %s %#v", me.Direction, r)
	return r > 0 && r <= dis
}
