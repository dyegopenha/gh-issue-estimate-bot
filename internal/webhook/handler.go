package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dyegopenha/gh-issue-estimate-bot/internal/estimate"
	"github.com/dyegopenha/gh-issue-estimate-bot/internal/githubapp"
	"github.com/google/go-github/v66/github"
)

const reminderMarker = "<!-- estimate-reminder -->"

type Handler struct {
	Logger        *log.Logger
	WebhookSecret string
}

func NewHandler(logger *log.Logger, secret string) *Handler {
	return &Handler{
		Logger:        logger,
		WebhookSecret: secret,
	}
}

// Register mounts the webhook endpoint to the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/webhook", h.serveWebhook)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func (h *Handler) serveWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature
	if !h.verifySignature(r.Header.Get("X-Hub-Signature-256"), body) {
		http.Error(w, "signature verification failed", http.StatusUnauthorized)
		return
	}

	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "issues" {
		// We ignore other events to keep surface small
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ignored"))
		return
	}

	var evt github.IssuesEvent
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&evt); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}

	if evt.GetAction() != "opened" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("non-opened action ignored"))
		return
	}

	issue := evt.GetIssue()
	repo := evt.GetRepo()
	inst := evt.GetInstallation()

	issueBody := issue.GetBody()
	author := issue.GetUser().GetLogin()
	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	issueNumber := issue.GetNumber()
	installationID := inst.GetID()

	h.Logger.Printf("issues.opened: %s/%s#%d by @%s", owner, repoName, issueNumber, author)

	if estimate.HasEstimate(issueBody) {
		h.Logger.Printf("estimate found; no action needed")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("estimate present"))
		return
	}

	// Create GitHub client for the installation
	cli, err := githubapp.NewInstallationClient(installationID)
	if err != nil {
		h.Logger.Printf("failed to create installation client: %v", err)
		http.Error(w, "failed to auth", http.StatusInternalServerError)
		return
	}

	// Idempotency: check for an existing reminder comment (marker)
	lastDays := time.Now().Add(-30 * 24 * time.Hour)
	comments, _, err := cli.Issues.ListComments(r.Context(), owner, repoName, issueNumber, &github.IssueListCommentsOptions{
		Since: &lastDays, // lookup last 30 days (cheap guard)
	})
	if err == nil {
		for _, c := range comments {
			if strings.Contains(c.GetBody(), reminderMarker) {
				h.Logger.Printf("existing reminder found; skipping new comment")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("reminder already posted"))
				return
			}
		}
	}

	// Post reminder
	comment := &github.IssueComment{
		Body: github.String(fmt.Sprintf(`%sHi @%s! Please add an estimate in the format **"Estimate: X days"** to help us plan and schedule work. Thanks!`, reminderMarker, author)),
	}
	_, _, err = cli.Issues.CreateComment(r.Context(), owner, repoName, issueNumber, comment)
	if err != nil {
		h.Logger.Printf("failed to create comment: %v", err)
		http.Error(w, "failed to comment", http.StatusInternalServerError)
		return
	}

	h.Logger.Printf("reminder comment posted on %s/%s#%d", owner, repoName, issueNumber)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("comment posted"))
}

func (h *Handler) verifySignature(sigHeader string, body []byte) bool {
	secret := h.WebhookSecret
	if secret == "" {
		// If not set, be explicit: deny
		return false
	}

	if !strings.HasPrefix(sigHeader, "sha256=") {
		return false
	}
	sigHex := strings.TrimPrefix(sigHeader, "sha256=")
	expected := hmac.New(sha256.New, []byte(secret))
	expected.Write(body)
	expectedSum := expected.Sum(nil)

	got, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}

	// Constant time compare
	if len(got) != len(expectedSum) {
		return false
	}
	if subtle.ConstantTimeCompare(got, expectedSum) != 1 {
		return false
	}
	return true
}
