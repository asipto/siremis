package main

import (
	"time"
)

type GMSession struct {
	username string
	expiry   time.Time
}

func (s GMSession) IsExpired() bool {
	return s.expiry.Before(time.Now())
}
