package main

import (
	"context"
	"errors"
)

// ErrContextCancelled is returned when the context is cancelled
var ErrContextCancelled = errors.New("context cancelled")

type uLoc struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

type visitID struct {
	VisitID string `json:"visitId"`
}

type uLocVisit struct {
	uLoc
	visitID
}

// HistoryManager interface exposes the core functionality of the data store
type HistoryManager interface {
	WriteHistory(context.Context, uLoc) (int, error)
	GetHistoryByVisitID(context.Context, string) ([]uLocVisit, error)
	GetHistoryByUserID(context.Context, string, string) ([]uLocVisit, error)
	Close(context.Context)
}
