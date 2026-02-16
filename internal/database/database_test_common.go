package database

import "sync"

// testDBMu serializes tests that use the global DB (Init/Close/CreateTask/etc.)
// so parallel tests do not overwrite DB and cause FK or "task not found" failures.
var testDBMu sync.Mutex
