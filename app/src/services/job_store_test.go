package services

import (
	"sync"
	"testing"
	"time"

	"github.dhi13man.com/bombardment-runner/src/models/dto"
)

func TestJobStore_Create(t *testing.T) {
	store := NewJobStore()
	job := store.Create(nil)

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
	job := store.Create(nil)

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
		job := store.Create(nil)
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
	job := store.Create(nil)

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
	job := store.Create(nil)

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
	job := store.Create(nil)
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
	job := store.Create(nil)
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

func TestJob_IncrementFailed(t *testing.T) {
	store := NewJobStore()
	job := store.Create(nil)
	job.SetRunning()
	job.SetTotal(10)

	for i := 0; i < 3; i++ {
		job.IncrementFailed()
	}

	snap := job.Snapshot()
	if snap.FailedRows != 3 {
		t.Fatalf("expected 3 failed rows, got %d", snap.FailedRows)
	}
}

func TestJob_ProgressPercent_IncludesFailedRows(t *testing.T) {
	store := NewJobStore()
	job := store.Create(nil)
	job.SetRunning()
	job.SetTotal(100)

	// 30 processed + 20 failed = 50 out of 100 = 50%
	for i := 0; i < 30; i++ {
		job.IncrementProcessed()
	}
	for i := 0; i < 20; i++ {
		job.IncrementFailed()
	}

	snap := job.Snapshot()
	const expected = 50.0
	const epsilon = 0.001
	if snap.ProgressPercent < expected-epsilon || snap.ProgressPercent > expected+epsilon {
		t.Fatalf("expected ProgressPercent ~%.1f (processed+failed), got %.4f", expected, snap.ProgressPercent)
	}
	if snap.ProcessedRows != 30 {
		t.Fatalf("expected 30 processed rows, got %d", snap.ProcessedRows)
	}
	if snap.FailedRows != 20 {
		t.Fatalf("expected 20 failed rows, got %d", snap.FailedRows)
	}
}

func TestJob_ProgressPercent_ZeroTotal(t *testing.T) {
	store := NewJobStore()
	job := store.Create(nil)
	job.SetRunning()
	// total stays at 0

	snap := job.Snapshot()
	if snap.ProgressPercent != 0 {
		t.Fatalf("expected ProgressPercent 0 when total is 0, got %.4f", snap.ProgressPercent)
	}
}

func TestJob_ConcurrentMixedUpdates(t *testing.T) {
	store := NewJobStore()
	job := store.Create(nil)
	job.SetRunning()
	job.SetTotal(200)

	var wg sync.WaitGroup
	wg.Add(200)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			job.IncrementProcessed()
		}()
		go func() {
			defer wg.Done()
			job.IncrementFailed()
		}()
	}
	wg.Wait()

	snap := job.Snapshot()
	if snap.ProcessedRows != 100 {
		t.Fatalf("expected 100 processed rows, got %d", snap.ProcessedRows)
	}
	if snap.FailedRows != 100 {
		t.Fatalf("expected 100 failed rows, got %d", snap.FailedRows)
	}

	const expectedProgress = 100.0
	const epsilon = 0.001
	if snap.ProgressPercent < expectedProgress-epsilon || snap.ProgressPercent > expectedProgress+epsilon {
		t.Fatalf("expected ProgressPercent ~%.1f, got %.4f", expectedProgress, snap.ProgressPercent)
	}
}

func TestJobStore_Delete_Exists(t *testing.T) {
	store := NewJobStore()
	job := store.Create(nil)

	ok := store.Delete(job.ID)
	if !ok {
		t.Fatal("expected Delete to return true for existing job")
	}

	_, found := store.Get(job.ID)
	if found {
		t.Fatal("expected Get to return false after deletion")
	}

	jobs := store.List()
	if len(jobs) != 0 {
		t.Fatalf("expected empty list after deletion, got %d", len(jobs))
	}
}

func TestJobStore_Delete_NotFound(t *testing.T) {
	store := NewJobStore()

	ok := store.Delete("nonexistent-id")
	if ok {
		t.Fatal("expected Delete to return false for non-existent ID")
	}
}

func TestJobStore_List_OmitsOriginalRequest(t *testing.T) {
	store := NewJobStore()
	req := &dto.BombardmentRequest{}
	req.Parser.FilePath = "/tmp/test.csv"
	store.Create(req)

	jobs := store.List()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].OriginalRequest != nil {
		t.Fatal("expected OriginalRequest to be nil in list results")
	}

	// Full snapshot via Get() should still include it
	snap, ok := store.Get(jobs[0].ID)
	if !ok {
		t.Fatal("expected Get to return true")
	}
	if snap.OriginalRequest == nil {
		t.Fatal("expected OriginalRequest to be present in Get() result")
	}
}

func TestJobStore_Create_StripsFileContentB64(t *testing.T) {
	store := NewJobStore()
	req := &dto.BombardmentRequest{}
	req.Parser.FileContentB64 = "dGVzdA==" // base64("test")
	req.Parser.FilePath = "/tmp/test.csv"

	job := store.Create(req)
	snap := job.Snapshot()

	if snap.OriginalRequest == nil {
		t.Fatal("expected OriginalRequest to be present")
	}
	if snap.OriginalRequest.Parser.FileContentB64 != "" {
		t.Fatal("expected file_content_b64 to be stripped from stored request")
	}
	if snap.OriginalRequest.Parser.FilePath != "/tmp/test.csv" {
		t.Fatalf("expected file_path to be preserved, got %q", snap.OriginalRequest.Parser.FilePath)
	}

	// Verify the original request was not mutated
	if req.Parser.FileContentB64 != "dGVzdA==" {
		t.Fatal("expected original request to remain unmutated")
	}
}

func TestJobStore_List_SortedNewestFirst(t *testing.T) {
	store := NewJobStore()

	job1 := store.Create(nil)
	time.Sleep(2 * time.Millisecond) // ensure distinct timestamps
	job2 := store.Create(nil)
	time.Sleep(2 * time.Millisecond)
	job3 := store.Create(nil)

	jobs := store.List()
	if len(jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(jobs))
	}
	// Newest first: job3, job2, job1
	if jobs[0].ID != job3.ID {
		t.Errorf("expected first job (newest) to be %q, got %q", job3.ID, jobs[0].ID)
	}
	if jobs[1].ID != job2.ID {
		t.Errorf("expected second job to be %q, got %q", job2.ID, jobs[1].ID)
	}
	if jobs[2].ID != job1.ID {
		t.Errorf("expected third job (oldest) to be %q, got %q", job1.ID, jobs[2].ID)
	}
}
