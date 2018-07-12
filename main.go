package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

const ROW = 10
const COL = 10

type SeatHold struct {
	seatHoldId  string
	heldSeats   []SeatInfo
	relatedInfo string
	expiration  <-chan time.Time
}

type SeatInfo struct {
	row, col int
	distance float64
}

// By is the type of a "less" function that defines the ordering of its SeatInfo arguments.
type By func(s1, s2 *SeatInfo) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(seatInfos []SeatInfo) {
	ss := &seatSorter{
		seatInfos: seatInfos,
		by:        by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ss)
}

// planetSorter joins a By function and a slice of Planets to be sorted.
type seatSorter struct {
	seatInfos []SeatInfo
	by        func(p1, p2 *SeatInfo) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *seatSorter) Len() int {
	return len(s.seatInfos)
}

// Swap is part of sort.Interface.
func (s *seatSorter) Swap(i, j int) {
	s.seatInfos[i], s.seatInfos[j] = s.seatInfos[j], s.seatInfos[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *seatSorter) Less(i, j int) bool {
	return s.by(&s.seatInfos[i], &s.seatInfos[j])
}

var (
	venue          [][]string
	readyToReserve map[string]SeatHold
)

func init() {
	venue = CreateVenue(ROW, COL)
	readyToReserve = make(map[string]SeatHold)
}

func main() {
	fmt.Println("Welcome to ticket service")
	exit := 0
	var input string
	var seatNum int
	var email string
	var holdID string
	for exit < 1 {
		// print venue
		for i := 0; i < len(venue); i++ {
			fmt.Println(venue[i])
		}
		fmt.Println("Options are: count seats available(count), hold seats(hold), reserve held seats(reserve), exit(exit)")
		fmt.Println("What would you like to do?:")
		fmt.Scanln(&input)
		switch input {
		case "count":
			fmt.Println("Number of available Seats are: ", numSeatsAvailable())
		case "hold":
			fmt.Println("How many seats do you need?")
			fmt.Scanln(&seatNum)
			fmt.Println("What's your email?")
			fmt.Scanln(&email)
			s := findAndHoldSeats(seatNum, email)
			fmt.Println("Your reservation ID is:", s.seatHoldId)
			// run concurrently and listen for expiration tick. When the expiration hits, it releases the held seat
			go func() {
				for {
					select {
					case <-s.expiration:
						// when it expires, it checks if it's already reserved. If it's reserved, nothing to release back
						seat, ok := readyToReserve[s.seatHoldId]
						if ok {
							delete(readyToReserve, seat.seatHoldId)
							for _, seatsOnHold := range seat.heldSeats {
								venue[seatsOnHold.row][seatsOnHold.col] = "O"
							}
						}

					}
				}
			}()
		case "reserve":
			//ask for reserveID and reserve
			fmt.Println("What's the hold ID?")
			fmt.Scanln(&holdID)
			fmt.Println("What's your email?")
			fmt.Scanln(&email)
			confirmation := reserveSeats(holdID, email)
			fmt.Println("Your reservation confirmation code is:", confirmation)
		case "exit":
			exit = 1
		default:
			fmt.Println("no correct input")
		}
	}
}

// CreateVenue : creates a venue of given size
func CreateVenue(row int, col int) [][]string {
	venue = make([][]string, row)
	for i := 0; i < row; i++ {
		venue[i] = make([]string, col)
		for j := 0; j < col; j++ {
			venue[i][j] = "O"
		}
	}
	return venue
}

// numSeatsAvailable: returns the number of seats avail
func numSeatsAvailable() int {
	var availableSeat int
	for _, row := range venue {
		for _, seat := range row {
			if seat == "O" {
				availableSeat++
			}
		}
	}
	return availableSeat
}

// findAndHoldSeats: takes number of seats customers are trying to book, and customer email
// and return the held seat object
func findAndHoldSeats(numSeats int, customerEmail string) SeatHold {
	// If it's not an int, it rounds down to the nearest int
	bestSeatRow := int(ROW / 2)
	bestSeatCol := int(COL / 2)
	// heldSeat := new(SeatHold)
	var heldSeat SeatHold
	distanceToBestSeat := func(s1, s2 *SeatInfo) bool {
		return s1.distance < s2.distance
	}
	for rowKey, row := range venue {
		for colKey, seat := range row {
			if seat == "O" {
				distance := math.Sqrt(math.Pow(float64(bestSeatRow-rowKey), 2) + math.Pow(float64(bestSeatCol-colKey), 2))
				currentSeat := SeatInfo{rowKey, colKey, distance}
				heldSeat.heldSeats = append(heldSeat.heldSeats, currentSeat)
			}
			if len(heldSeat.heldSeats) > numSeats {
				By(distanceToBestSeat).Sort(heldSeat.heldSeats)
				// deleting array to leave the least distance to the best seat
				heldSeat.heldSeats = append(heldSeat.heldSeats[:numSeats])
			}
		}
	}
	// set the marker to show the held seats
	for _, seatsOnHold := range heldSeat.heldSeats {
		venue[seatsOnHold.row][seatsOnHold.col] = "H"
	}
	// assigning random reservation number
	randNum := rand.Intn(99999)
	id := PadLeft(strconv.Itoa(randNum), "0", 6)
	heldSeat.seatHoldId = id
	heldSeat.relatedInfo = customerEmail
	//after 10 seconds, hold seats will be released
	heldSeat.expiration = time.After(10000 * time.Millisecond)
	readyToReserve[id] = heldSeat
	return heldSeat
}

// PadLeft : is used to fill up and make sure all reserveID are specified length
func PadLeft(str, pad string, lenght int) string {
	for {
		str = pad + str
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

// reserveSeats: takes hold identifier, and customer email, and return confirmation string.
// If reserveSeats failed, return 'RESERVATION_FAILED'
func reserveSeats(seatHoldId string, customerEmail string) string {
	seat, ok := readyToReserve[seatHoldId]
	if ok {
		for _, seatsOnHold := range seat.heldSeats {
			fmt.Println(venue[seatsOnHold.row][seatsOnHold.col])
			venue[seatsOnHold.row][seatsOnHold.col] = "R"
			delete(readyToReserve, seatHoldId)
		}
		return "Random confirmation string"
	} else {
		return "RESERVATION_FAILED. Hold seat ID does not exist"
	}
}
