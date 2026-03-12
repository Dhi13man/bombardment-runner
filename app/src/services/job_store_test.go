package services

import (
	"sync"
	"testing"
	"time"
)

func TestJobStore_Create(t *testing.T) {
	store := NewJobStore()
	job := store.Create()

	if job.ID == "" {
		t.Fatal("expected non-empty UUID, got empty string")
	}
	if job.Status != JobStatusPending {
		t.Fatalf("expected status %q, got %q", JobStatusPending, job.Status)
	}
	if job.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
}

func TestJobStore_Get_Exists(t *testing.T) {
	store := NewJobStore()
	job := store.Create()

	snap, ok := store.Get(job.ID)
	if !ok {
		t.Fatal("expected Get to return true for existing job")
	}
	if snap.ID != job.ID {
		t.Fatalf("expected snapshot ID %q, got %q", job.ID, snap.ID)
	}
	if snap.Status != JobStatusPending {
		t.Fatalf("expected status %q, got %q", JobStatusPending, snap.Status)
	}
}

func TestJobStore_Get_NotFound(t *testing.T) {
	store := NewJobStore()

	_, ok := store.Get("nonexistent-id")
	if ok {
		t.Fatal("expected Get to return false for non-existent ID")
	}
}

func TestJobStore_List_Empty(t *testing.T) {
	store := NewJobStore()

	jobs := store.List()
	if len(jobs) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(jobs))
	}
}

func TestJobStore_List_Multiple(t *testing.T) {
	store := NewJobStore()

	ids := make(map[string]bool)
	for i := 0; i < 3; i++ {
		job := store.Create()
		ids[job.ID] = true
	}

	jobs := store.List()
	if len(jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(jobs))
	}

	for _, snap := range jobs {
		if !ids[snap.ID] {
			t.Fatalf("unexpected job ID %q in list", snap.ID)
		}
	}
}

func TestJob_Lifecycle_PendingToCompleted(t *testing.T) {
	store := NewJobStore()
	job := store.Create()

	// Verify initial state
	if job.Status != JobStatusPending {
		t.Fatalf("expected initial status %q, got %q", JobStatusPending, job.Status)
	}

	// Transition to running
	job.SetRunning()
	snap := job.Snapshot()
	if snap.Status != JobStatusRunning {
		t.Fatalf("expected status %q after SetRunning, got %q", JobStatusRunning, snap.Status)
	}

	// Set total rows
	job.SetTotal(10)

	// Process 5 rows
	for i := 0; i < 5; i++ {
		job.IncrementProcessed()
	}

	snap = job.Snapshot()
	if snap.ProcessedRows != 5 {
		t.Fatalf("expected 5 processed rows, got %d", snap.ProcessedRows)
	}
	if snap.TotalRows != 10 {
		t.Fatalf("expected 10 total rows, got %d", snap.TotalRows)
	}

	// Complete the job
	beforeComplete := time.Now()
	job.Complete()
	snap = job.Snapshot()

	if snap.Status != JobStatusCompleted {
		t.Fatalf("expected status %q after Complete, got %q", JobStatusCompleted, snap.Status)
	}
	if snap.CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set after Complete")
	}
	if snap.CompletedAt.Before(beforeComplete) {
		t.Fatal("expected CompletedAt to be at or after the time Complete was called")
	}
}

func TestJob_Lifecycle_PendingToFailed(t *testing.T) {
	store := NewJobStore()
	job := store.Create()

	job.SetRunning()

	beforeFail := time.Now()
	job.Fail("boom")
	snap := job.Snapshot()

	if snap.Status != JobStatusFailed {
		t.Fatalf("expected status %q, got %q", JobStatusFailed, snap.Status)
	}
	if snap.ErrorMessage != "boom" {
		t.Fatalf("expected error message %q, got %q", "boom", snap.ErrorMessage)
	}
	if snap.CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set after Fail")
	}
	if snap.CompletedAt.Before(beforeFail) {
		t.Fatal("expected CompletedAt to be at or after the time Fail was called")
	}
}

func TestJob_ConcurrentUpdates(t *testing.T) {
	store := NewJobStore()
	job := store.Create()
	job.SetRunning()
	job.SetTotal(100)

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			job.IncrementProcessed()
		}()
	}
	wg.Wait()

	snap := job.Snapshot()
	if snap.ProcessedRows != 100 {
		t.Fatalf("expected 100 processed rows after concurrent updates, got %d", snap.ProcessedRows)
	}
}

func TestJob_ProgressPercent(t *testing.T) {
	store := NewJobStore()
	job := store.Create()
	job.SetRunning()
	job.SetTotal(100)

	for i := 0; i < 50; i++ {
		job.IncrementProcessed()
	}

	snap := job.Snapshot()

	// ProgressPercent = (processed + failed) / total * 100
	// With 50 processed, 0 failed, 100 total: expect 50.0
	const expected = 50.0
	const epsilon = 0.001
	if snap.ProgressPercent < expected-epsilon || snap.ProgressPercent > expected+epsilon {
		t.Fatalf("expected ProgressPercent ~%.1f, got %.4f", expected, snap.ProgressPercent)
	}
}
