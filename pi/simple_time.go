package main

import (
	"strings"
	"time"
)

type SimpleTime struct {
	time.Time
}

func (st *SimpleTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	s = strings.ToLower(s)

	var layout string
	if strings.Contains(s, ":") {
		layout = "3:04pm"
	} else {
		layout = "3pm"
	}

	t, err := time.Parse(layout, s)
	if err != nil {
		return err
	}

	st.Time = t
	return nil
}

type Schedule struct {
	Event string     `json:"event"`
	Time  SimpleTime `json:"time"`
}
