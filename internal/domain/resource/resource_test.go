package resource

import "testing"

func TestNewFileResourceValidatesInputs(t *testing.T) {
	t.Run("missing group", func(t *testing.T) {
		if _, err := NewFile("", "notes.pdf"); err == nil {
			t.Fatalf("expected missing group error")
		}
	})

	t.Run("unsupported extension", func(t *testing.T) {
		if _, err := NewFile("group-1", "archive.zip"); err == nil {
			t.Fatalf("expected unsupported extension error")
		}
	})
}

func TestNewWebResourceValidatesURL(t *testing.T) {
	if _, err := NewWeb("group-1", "not-a-url"); err == nil {
		t.Fatalf("expected invalid url error")
	}
}

func TestStatusMachineAllowsExpectedTransitions(t *testing.T) {
	entity, err := NewFile("group-1", "guide.md")
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}

	for _, next := range []Status{StatusParsing, StatusNormalizing, StatusGraphGenerating, StatusCompleted} {
		if err := entity.MoveTo(next, Failure{}); err != nil {
			t.Fatalf("MoveTo(%s) error = %v", next, err)
		}
	}
}

func TestStatusMachineRejectsIllegalTransition(t *testing.T) {
	entity, err := NewFile("group-1", "guide.md")
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}

	if err := entity.MoveTo(StatusCompleted, Failure{}); err == nil {
		t.Fatalf("expected invalid transition")
	}
}

func TestFailTransitionRequiresStageAndMessage(t *testing.T) {
	entity, err := NewFile("group-1", "guide.md")
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}

	if err := entity.MoveTo(StatusFailed, Failure{}); err == nil {
		t.Fatalf("expected failure metadata validation")
	}

	if err := entity.MoveTo(StatusFailed, Failure{Stage: StatusParsing, Message: "tika timeout"}); err != nil {
		t.Fatalf("MoveTo(failed) error = %v", err)
	}
	if entity.Failure == nil || entity.Failure.Message != "tika timeout" {
		t.Fatalf("failure metadata was not saved")
	}
}

func TestRetryResetsFailedOrCompletedResources(t *testing.T) {
	failed, err := NewFile("group-1", "guide.md")
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}
	if err := failed.MoveTo(StatusFailed, Failure{Stage: StatusParsing, Message: "bad file"}); err != nil {
		t.Fatalf("MoveTo(failed) error = %v", err)
	}
	if err := failed.ResetForRetry(); err != nil {
		t.Fatalf("ResetForRetry() error = %v", err)
	}
	if failed.Status != StatusUploaded {
		t.Fatalf("status = %s, want uploaded", failed.Status)
	}
	if failed.Failure != nil {
		t.Fatalf("failure should be cleared")
	}

	completed, err := NewWeb("group-1", "https://example.com/page")
	if err != nil {
		t.Fatalf("NewWeb() error = %v", err)
	}
	for _, next := range []Status{StatusParsing, StatusNormalizing, StatusGraphGenerating, StatusCompleted} {
		if err := completed.MoveTo(next, Failure{}); err != nil {
			t.Fatalf("MoveTo(%s) error = %v", next, err)
		}
	}
	if err := completed.ResetForRetry(); err != nil {
		t.Fatalf("ResetForRetry() error = %v", err)
	}
	if completed.Status != StatusUploaded {
		t.Fatalf("status = %s, want uploaded", completed.Status)
	}
}
