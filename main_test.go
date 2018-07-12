package main

import "testing"

func TestReserveSeats(t *testing.T) {
	confirmation := reserveSeats("whatever", "sample email")
	if confirmation != "RESERVATION_FAILED. Hold seat ID does not exist" {
		t.Errorf("It should have failed")
	}
	// write a successful test case
	heldSeatObj := findAndHoldSeats(2, "sample email")
	if heldSeatObj.seatHoldId != "" {
		confirmation := reserveSeats(heldSeatObj.seatHoldId, "sample email")
		if confirmation != "Random confirmation string" {
			t.Errorf("successful reservation test case failed")
		}
	}
}

func TestFindAndHoldSeats(t *testing.T) {
	heldSeatObj := findAndHoldSeats(2, "sample email")
	if heldSeatObj.seatHoldId == "" {
		t.Errorf("It should have created a reservation")
	}
}
