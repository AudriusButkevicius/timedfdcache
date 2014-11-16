timedfdcache
============

A time based file descriptor cache.

When `Close()` is called on a file, the closure of the file gets deferred for the configured period of time.
During that time, any `Open()` call for the same path will instead return the same file descriptor which is deferred for
closure cancelling the closure.
