package GitDb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitDbDurability_CloseReopen_SeesAllKeys(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitdb-durability-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db1, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open(db1): %v", err)
	}
	if err := db1.Put("key1", []byte("value1")); err != nil {
		t.Fatalf("Put(key1): %v", err)
	}
	if err := db1.Close(); err != nil {
		t.Fatalf("Close(db1): %v", err)
	}

	db2, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open(db2): %v", err)
	}
	v1, err := db2.Get("key1")
	if err != nil {
		t.Fatalf("Get(key1) after reopen: %v", err)
	}
	if string(v1) != "value1" {
		t.Fatalf("unexpected value for key1: %q", string(v1))
	}
	if err := db2.Put("key2", []byte("value2")); err != nil {
		t.Fatalf("Put(key2): %v", err)
	}
	if err := db2.Close(); err != nil {
		t.Fatalf("Close(db2): %v", err)
	}

	db3, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open(db3): %v", err)
	}
	defer db3.Close()

	v1b, err := db3.Get("key1")
	if err != nil {
		t.Fatalf("Get(key1) after second reopen: %v", err)
	}
	if string(v1b) != "value1" {
		t.Fatalf("unexpected value for key1 after second reopen: %q", string(v1b))
	}

	v2, err := db3.Get("key2")
	if err != nil {
		t.Fatalf("Get(key2) after second reopen: %v", err)
	}
	if string(v2) != "value2" {
		t.Fatalf("unexpected value for key2 after second reopen: %q", string(v2))
	}
}

// This repro intentionally uses 2 overlapping handles. The older handle must NOT be able
// to drop the newer handle's appended records when it closes.
//
// This should FAIL on the current buggy implementation where Close() rewrites/truncates
// the log file from the handle's in-memory snapshot.
func TestGitDbDurability_CloseMustNotTruncateOrDropNewerAppends_MultiHandle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitdb-durability-multihandle-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "log")

	// Handle A writes keyA
	handleA, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open(handleA): %v", err)
	}
	if err := handleA.Put("keyA", []byte("A")); err != nil {
		t.Fatalf("handleA.Put(keyA): %v", err)
	}

	// Handle B opens after keyA exists, then appends keyB
	handleB, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open(handleB): %v", err)
	}
	if _, err := handleB.Get("keyA"); err != nil {
		t.Fatalf("handleB.Get(keyA): %v", err)
	}
	if err := handleB.Put("keyB", []byte("B")); err != nil {
		t.Fatalf("handleB.Put(keyB): %v", err)
	}

	fiBefore, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat before Close(handleA): %v", err)
	}

	if err := handleB.Close(); err != nil {
		t.Fatalf("Close(handleB): %v", err)
	}

	// Closing stale handleA must not truncate/overwrite and drop keyB.
	if err := handleA.Close(); err != nil {
		t.Fatalf("Close(handleA): %v", err)
	}

	fiAfter, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat after Close(handleA): %v", err)
	}

	if fiAfter.Size() < fiBefore.Size() {
		t.Fatalf("log file shrank after Close(handleA): before=%d after=%d (truncation detected)", fiBefore.Size(), fiAfter.Size())
	}

	handleC, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open(handleC): %v", err)
	}
	defer handleC.Close()

	if _, err := handleC.Get("keyA"); err != nil {
		t.Fatalf("handleC.Get(keyA): %v", err)
	}
	if _, err := handleC.Get("keyB"); err != nil {
		t.Fatalf("handleC.Get(keyB): %v (keyB was dropped)", err)
	}
}

func TestGitDbDurability_RemoteRef_CloseReopen(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitdb-durability-remote-ref-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	key := "refs/remotes/origin/master"
	val := []byte("commit123")

	db1, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open(db1): %v", err)
	}
	if err := db1.Put(key, val); err != nil {
		t.Fatalf("Put(remoteRef): %v", err)
	}
	if err := db1.Close(); err != nil {
		t.Fatalf("Close(db1): %v", err)
	}

	db2, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open(db2): %v", err)
	}
	defer db2.Close()

	got, err := db2.Get(key)
	if err != nil {
		t.Fatalf("Get(remoteRef) after reopen: %v", err)
	}
	if string(got) != string(val) {
		t.Fatalf("unexpected remoteRef value: got=%q want=%q", string(got), string(val))
	}
}



