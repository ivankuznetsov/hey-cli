package smoke_test

import (
	"fmt"
	"testing"
)

func TestPostingActions(t *testing.T) {
	actions := []struct {
		command     string
		pastTense   string
		destination string
	}{
		{command: "paper-trail", pastTense: "moved to Paper Trail", destination: "trailbox"},
		{command: "feed", pastTense: "moved to The Feed", destination: "feedbox"},
		{command: "set-aside", pastTense: "moved to Set Aside", destination: "asidebox"},
		{command: "reply-later", pastTense: "moved to Reply Later", destination: "laterbox"},
		{command: "ignore", pastTense: "ignored"},
		{command: "trash", pastTense: "moved to Trash"},
	}

	for _, action := range actions {
		t.Run(action.command, func(t *testing.T) {
			postingID := createPostingForAction(t)
			postingIDString := intStr(postingID)
			t.Cleanup(func() {
				hey(t, "trash", postingIDString, "--kind", "topic", "--json")
			})

			resp := heyJSON(t, action.command, postingIDString, "--kind", "topic")
			wantSummary := fmt.Sprintf("Posting %d %s", postingID, action.pastTense)
			if resp.Summary != wantSummary {
				t.Errorf("summary = %q, want %q", resp.Summary, wantSummary)
			}
			if postingInBox(t, "imbox", postingID) {
				t.Errorf("posting %d still present in imbox after %s", postingID, action.command)
			}
			if action.destination != "" && !postingInBox(t, action.destination, postingID) {
				t.Errorf("posting %d not found in %s after %s", postingID, action.destination, action.command)
			}
		})
	}
}

func createPostingForAction(t *testing.T) int {
	t.Helper()
	subject := fmt.Sprintf("Planning follow-up %s", uniqueID())
	_, stderr, code := hey(t,
		"compose",
		"--to", accountEmail,
		"--subject", subject,
		"--message", "Agenda and notes for the upcoming planning review.",
		"--json",
	)
	if code != 0 {
		t.Skipf("could not create a posting for action smoke coverage (exit %d): %s", code, stderr)
	}

	resp := heyJSON(t, "box", "imbox", "--all")
	type Posting struct {
		ID      int    `json:"id"`
		Summary string `json:"summary"`
	}
	type BoxResp struct {
		Postings []Posting `json:"postings"`
	}
	data := dataAs[BoxResp](t, resp)
	for _, posting := range data.Postings {
		if posting.Summary == subject {
			return posting.ID
		}
	}

	t.Fatalf("could not find composed posting %q in imbox", subject)
	return 0
}

func postingInBox(t *testing.T, box string, postingID int) bool {
	t.Helper()
	resp := heyJSON(t, "box", box, "--all")
	type Posting struct {
		ID int `json:"id"`
	}
	type BoxResp struct {
		Postings []Posting `json:"postings"`
	}
	data := dataAs[BoxResp](t, resp)
	for _, posting := range data.Postings {
		if posting.ID == postingID {
			return true
		}
	}
	return false
}
