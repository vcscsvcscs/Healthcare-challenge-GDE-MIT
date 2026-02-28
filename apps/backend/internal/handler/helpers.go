package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

// Helper functions for type conversions between API types and internal models

// stringPtr creates a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// intPtr creates a pointer to an int
func intPtr(i int) *int {
	return &i
}

// boolPtr creates a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}

// timePtr creates a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}

// uuidToString converts types.UUID to string
func uuidToString(u types.UUID) string {
	return uuid.UUID(u).String()
}

// stringToUUID converts string to types.UUID pointer
func stringToUUID(s string) *types.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	apiUUID := types.UUID(u)
	return &apiUUID
}

// dateToTime converts types.Date to time.Time
func dateToTime(d types.Date) time.Time {
	return d.Time
}

// timeToDate converts time.Time to types.Date pointer
func timeToDate(t time.Time) *types.Date {
	return &types.Date{Time: t}
}

// timePtrToDate converts *time.Time to *types.Date
func timePtrToDate(t *time.Time) *types.Date {
	if t == nil {
		return nil
	}
	return &types.Date{Time: *t}
}
