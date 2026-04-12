package policies_test

import (
	"testing"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
	"github.com/yolo-hq/yolo/yolotest"
)

// CanAcknowledgeInsight

func TestCanAcknowledgeInsight_AllowPending(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanAcknowledgeInsightPolicy{},
		map[string]any{"status": "pending"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanAcknowledgeInsight_DenyNotPending(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanAcknowledgeInsightPolicy{},
		map[string]any{"status": "acknowledged"},
	)
	if result.Allowed {
		t.Fatal("expected denied for non-pending insight")
	}
}

// CanAnswerQuestion

func TestCanAnswerQuestion_AllowOpen(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanAnswerQuestionPolicy{},
		map[string]any{"status": "open"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanAnswerQuestion_DenyNotOpen(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanAnswerQuestionPolicy{},
		map[string]any{"status": "answered"},
	)
	if result.Allowed {
		t.Fatal("expected denied for answered question")
	}
}

// CanApplyInsight

func TestCanApplyInsight_AllowAcknowledged(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanApplyInsightPolicy{},
		map[string]any{"status": "acknowledged"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanApplyInsight_DenyNotAcknowledged(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanApplyInsightPolicy{},
		map[string]any{"status": "pending"},
	)
	if result.Allowed {
		t.Fatal("expected denied for pending insight")
	}
}

// CanApprovePRD

func TestCanApprovePRD_AllowDraft(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanApprovePRDPolicy{},
		map[string]any{"status": "draft"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanApprovePRD_DenyNotDraft(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanApprovePRDPolicy{},
		map[string]any{"status": "approved"},
	)
	if result.Allowed {
		t.Fatal("expected denied for non-draft PRD")
	}
}

// CanApproveSuggestion

func TestCanApproveSuggestion_AllowPending(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanApproveSuggestionPolicy{},
		map[string]any{"status": "pending"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanApproveSuggestion_DenyNotPending(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanApproveSuggestionPolicy{},
		map[string]any{"status": "approved"},
	)
	if result.Allowed {
		t.Fatal("expected denied for approved suggestion")
	}
}

// CanArchiveProject

func TestCanArchiveProject_AllowActive(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanArchiveProjectPolicy{},
		map[string]any{"status": "active"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanArchiveProject_DenyAlreadyArchived(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanArchiveProjectPolicy{},
		map[string]any{"status": "archived"},
	)
	if result.Allowed {
		t.Fatal("expected denied for already archived project")
	}
}

// CanCancelTask

func TestCanCancelTask_AllowQueued(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanCancelTaskPolicy{},
		map[string]any{"status": "queued"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanCancelTask_DenyDone(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanCancelTaskPolicy{},
		map[string]any{"status": "done"},
	)
	if result.Allowed {
		t.Fatal("expected denied for done task")
	}
}

func TestCanCancelTask_DenyFailed(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanCancelTaskPolicy{},
		map[string]any{"status": "failed"},
	)
	if result.Allowed {
		t.Fatal("expected denied for failed task")
	}
}

func TestCanCancelTask_DenyCancelled(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanCancelTaskPolicy{},
		map[string]any{"status": "cancelled"},
	)
	if result.Allowed {
		t.Fatal("expected denied for already cancelled task")
	}
}

// CanDismissInsight

func TestCanDismissInsight_AllowPending(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanDismissInsightPolicy{},
		map[string]any{"status": "pending"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed for pending, got: %s", result.Reason)
	}
}

func TestCanDismissInsight_AllowAcknowledged(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanDismissInsightPolicy{},
		map[string]any{"status": "acknowledged"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed for acknowledged, got: %s", result.Reason)
	}
}

func TestCanDismissInsight_DenyApplied(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanDismissInsightPolicy{},
		map[string]any{"status": "applied"},
	)
	if result.Allowed {
		t.Fatal("expected denied for applied insight")
	}
}

// CanExecutePRD

func TestCanExecutePRD_AllowDraft(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanExecutePRDPolicy{},
		map[string]any{"status": "draft"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed for draft, got: %s", result.Reason)
	}
}

func TestCanExecutePRD_AllowApproved(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanExecutePRDPolicy{},
		map[string]any{"status": "approved"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed for approved, got: %s", result.Reason)
	}
}

func TestCanExecutePRD_DenyCompleted(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanExecutePRDPolicy{},
		map[string]any{"status": "completed"},
	)
	if result.Allowed {
		t.Fatal("expected denied for completed PRD")
	}
}

// CanPauseProject

func TestCanPauseProject_AllowActive(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanPauseProjectPolicy{},
		map[string]any{"status": "active"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanPauseProject_DenyNotActive(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanPauseProjectPolicy{},
		map[string]any{"status": "paused"},
	)
	if result.Allowed {
		t.Fatal("expected denied for already paused project")
	}
}

// CanRejectSuggestion

func TestCanRejectSuggestion_AllowPending(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanRejectSuggestionPolicy{},
		map[string]any{"status": "pending"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanRejectSuggestion_DenyNotPending(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanRejectSuggestionPolicy{},
		map[string]any{"status": "rejected"},
	)
	if result.Allowed {
		t.Fatal("expected denied for already rejected suggestion")
	}
}

// CanResumeProject

func TestCanResumeProject_AllowPaused(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanResumeProjectPolicy{},
		map[string]any{"status": "paused"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanResumeProject_DenyNotPaused(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanResumeProjectPolicy{},
		map[string]any{"status": "active"},
	)
	if result.Allowed {
		t.Fatal("expected denied for active project")
	}
}

// CanRetryTask

func TestCanRetryTask_AllowFailed(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanRetryTaskPolicy{},
		map[string]any{"status": "failed"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanRetryTask_DenyNotFailed(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanRetryTaskPolicy{},
		map[string]any{"status": "done"},
	)
	if result.Allowed {
		t.Fatal("expected denied for non-failed task")
	}
}

// CanSubmitPRD

func TestCanSubmitPRD_AllowActiveProject(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanSubmitPRDPolicy{},
		map[string]any{"status": "active"},
	)
	if !result.Allowed {
		t.Fatalf("expected allowed, got: %s", result.Reason)
	}
}

func TestCanSubmitPRD_DenyNotActive(t *testing.T) {
	result := yolotest.EvaluateEntityPolicy(t, &policies.CanSubmitPRDPolicy{},
		map[string]any{"status": "paused"},
	)
	if result.Allowed {
		t.Fatal("expected denied for non-active project")
	}
}
