package timedfdcache

import (
	"io/ioutil"
	"os"
	"time"

	"testing"
)

// Check that closing and waiting for a given period of time actually closes.
func TestClosing(t *testing.T) {
	fd, err := ioutil.TempFile("", "fdcache")
	path := fd.Name()
	if err != nil {
		t.Fatal(err)
		return
	}
	fd.WriteString("test")
	fd.Close()
	buf := make([]byte, 4)
	defer os.Remove(path)

	c := NewCache(time.Millisecond * 10)
	f1, err := c.Open(path)
	if err != nil {
		t.Fatal(err)
		return
	}
	f1.Close()
	time.Sleep(time.Millisecond * 15)
	_, err = f1.Read(buf)
	if err == nil {
		t.Fatal("Expected error", err)
		return
	}
}

// Check that we still can read after closing
func TestDelayClosing(t *testing.T) {
	fd, err := ioutil.TempFile("", "fdcache")
	path := fd.Name()
	if err != nil {
		t.Fatal(err)
		return
	}
	fd.WriteString("test")
	fd.Close()
	buf := make([]byte, 4)
	defer os.Remove(path)

	c := NewCache(time.Second)
	f1, err := c.Open(path)
	if err != nil {
		t.Fatal(err)
		return
	}
	f1.Close()
	_, err = f1.Read(buf)
	if err != nil {
		t.Fatal("Unexpected error", err)
		return
	}
}

// Checks that concurrent opens to the same file doesn't leave fd's hanging
func TestConcurrentOpen(t *testing.T) {
	fd, err := ioutil.TempFile("", "fdcache")
	path := fd.Name()
	if err != nil {
		t.Fatal(err)
		return
	}
	fd.WriteString("test")
	fd.Close()
	buf := make([]byte, 4)
	defer os.Remove(path)

	c := NewCache(time.Millisecond * 10)
	f1, err := c.Open(path)
	if err != nil {
		t.Fatal(err)
		return
	}

	f2, err := c.Open(path)
	if err != nil {
		t.Fatal(err)
		return
	}

	f1.Close()
	f2.Close()

	f3, err := c.Open(path)
	if err != nil {
		t.Fatal(err)
		return
	}

	time.Sleep(time.Millisecond * 15)
	// At this point, f1 should be closed as it got overwritten by f2 and never
	// had it's timer cancelled, and f3 should be f2.

	_, err = f1.ReadAt(buf, 0)
	if err == nil {
		t.Fatal("Expected error")
		return
	}

	if f2 != f3 {
		t.Fatal("Unexpected fd")
	}

	_, err = f2.ReadAt(buf, 0)
	if err != nil {
		t.Fatal("Unexpected error", err)
		return
	}

	_, err = f3.ReadAt(buf, 0)
	if err != nil {
		t.Fatal("Unexpected error", err)
		return
	}
}
