package handlers

import (
	"net/url"
	"strconv"

	"github.com/google/uuid"
)

// typedParam reads a query parameter and converts it to a typed string alias (e.g. models.IssueStatus).
// Returns nil if the parameter is absent.
func typedParam[T ~string](params url.Values, key string) *T {
	if v := params.Get(key); v != "" {
		t := T(v)
		return &t
	}
	return nil
}

// queryParam reads a query parameter and returns a pointer to the value, or nil if absent.
func queryParam(params url.Values, key string) *string {
	if v := params.Get(key); v != "" {
		return &v
	}
	return nil
}

// queryParamBool reads a query parameter as a *bool ("true"/"false"), or nil if absent.
func queryParamBool(params url.Values, key string) *bool {
	if v := params.Get(key); v != "" {
		b := v == "true"
		return &b
	}
	return nil
}

// queryParamUUID reads a query parameter as a *uuid.UUID, or nil if absent or invalid.
func queryParamUUID(params url.Values, key string) *uuid.UUID {
	if v := params.Get(key); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			return &id
		}
	}
	return nil
}

// queryParamInt reads a query parameter as an int, returning 0 if absent or invalid.
func queryParamInt(params url.Values, key string) int {
	if v := params.Get(key); v != "" {
		n, _ := strconv.Atoi(v)
		return n
	}
	return 0
}
