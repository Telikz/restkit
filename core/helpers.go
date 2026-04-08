package core

import (
	ep "github.com/reststore/restkit/internal/endpoints"
	mw "github.com/reststore/restkit/internal/middleware"
)

// ParseID converts a string ID to int64.
var ParseID = ep.ParseID

// ParseIntID converts a string ID to int.
var ParseIntID = ep.ParseIntID

// ParseUUID converts a string ID to UUID [16]byte.
var ParseUUID = ep.ParseUUID

// StringToInt converts a string to int.
var StringToInt = mw.StringToInt

// StringToString is a no-op converter for string path params.
var StringToString = mw.StringToString
