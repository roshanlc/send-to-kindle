package helper

import (
	"os"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

const ctxTaskKey = "taskID"
const ctxFilepathKey = "filepath"

// GenerateID returns a new UUID
func GenerateID() uuid.UUID {
	return uuid.New()
}

// NewContextWithUUID creates a context containg the ID provided in the given context
func NewContextWithUUID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxTaskKey, id)
}

// GenerateIDWithContext creates a new context with a newly generated ID
func GenerateIDWithContext() context.Context {
	return context.WithValue(context.Background(), ctxTaskKey, uuid.New())
}

// GetIDFromContext retrieves the uuid from the provided context.
// Returns an empty uuid incase it does not exist.
// The receiver has the burden to check if the returned UUID is valid or not.
func GetIDFromContext(ctx context.Context) uuid.UUID {
	val, _ := ctx.Value(ctxTaskKey).(uuid.UUID)
	return val
}

// IsUUIDValid checks if the given ID is nil
func IsUUIDValid(id uuid.UUID) bool {
	return id == uuid.Nil
}

// NewContextWithFilePath creates a context having filepath value
func NewContextWithFilePath(ctx context.Context, filepath string) context.Context {
	return context.WithValue(ctx, ctxFilepathKey, filepath)
}

// GetFilepathFromContext retrieves filepath from given context.
// The receiver has the burden to check if the returned path is valid or not.
func GetFilepathFromContext(ctx context.Context) string {
	val, _ := ctx.Value(ctxFilepathKey).(string)
	return val
}

// IsFilepathValid checks if filepath is valid
func IsFilepathValid(filepath string) bool {
	if strings.TrimSpace(filepath) == "" {
		return false
	}

	if _, err := os.Stat(filepath); err != nil {
		return false
	}
	return true
}
