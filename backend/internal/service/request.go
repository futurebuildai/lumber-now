package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

type RequestService struct {
	store       *store.Store
	aiClient    AIParser
	transcriber Transcriber
	emailClient EmailSender
	mediaSvc    *MediaService
	emailWg     sync.WaitGroup
}

type AIParser interface {
	ParseText(ctx context.Context, text string, inventory []db.Inventory) ([]domain.StructuredItem, float64, error)
	ParseImage(ctx context.Context, imageURL string, inventory []db.Inventory) ([]domain.StructuredItem, float64, error)
	ParsePDF(ctx context.Context, pdfURL string, inventory []db.Inventory) ([]domain.StructuredItem, float64, error)
}

type Transcriber interface {
	Transcribe(ctx context.Context, audioData io.Reader) (string, error)
}

type EmailSender interface {
	SendOrderConfirmation(ctx context.Context, toEmail, dealerName string, items []domain.StructuredItem) error
}

func NewRequestService(s *store.Store, ai AIParser, transcriber Transcriber, emailClient EmailSender, mediaSvc *MediaService) *RequestService {
	return &RequestService{
		store:       s,
		aiClient:    ai,
		transcriber: transcriber,
		emailClient: emailClient,
		mediaSvc:    mediaSvc,
	}
}

// Close waits for all in-flight email goroutines to complete.
func (s *RequestService) Close() {
	s.emailWg.Wait()
}

// wrapVersionErr converts pgx.ErrNoRows from versioned queries into domain.ErrVersionConflict.
func wrapVersionErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrVersionConflict
	}
	return err
}

type CreateRequestInput struct {
	DealerID     uuid.UUID
	ContractorID uuid.UUID
	InputType    domain.InputType
	RawText      string
	MediaURL     string
}

func (s *RequestService) Create(ctx context.Context, input CreateRequestInput) (*db.Request, error) {
	if !input.InputType.Valid() {
		return nil, domain.ErrInvalidInput
	}

	// Auto-assign rep if contractor has one
	var assignedRepID pgtype.UUID
	contractor, err := s.store.Queries.GetUser(ctx, input.ContractorID)
	if err == nil && contractor.AssignedRepID.Valid {
		assignedRepID = contractor.AssignedRepID
	}

	req, err := s.store.Queries.CreateRequest(ctx, db.CreateRequestParams{
		DealerID:      input.DealerID,
		ContractorID:  input.ContractorID,
		AssignedRepID: assignedRepID,
		InputType:     db.InputType(input.InputType),
		RawText:       input.RawText,
		MediaUrl:      input.MediaURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	return &req, nil
}

func (s *RequestService) Process(ctx context.Context, requestID uuid.UUID) (*db.Request, error) {
	req, err := s.store.Queries.GetRequest(ctx, requestID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if req.Status != db.RequestStatusPending && req.Status != db.RequestStatusProcessing {
		return nil, domain.ErrInvalidStatus
	}

	// Set to processing if not already (worker pre-claims via ClaimPendingRequests)
	if req.Status == db.RequestStatusPending {
		req, err = s.store.Queries.UpdateRequestStatusVersioned(ctx, db.UpdateRequestStatusVersionedParams{
			ID:      requestID,
			Status:  db.RequestStatusProcessing,
			Version: req.Version,
		})
		if err != nil {
			return nil, fmt.Errorf("update status: %w", wrapVersionErr(err))
		}
	}

	// Get dealer inventory for SKU matching
	inventory, err := s.store.Queries.ListInventory(ctx, db.ListInventoryParams{
		DealerID: req.DealerID,
		Limit:    500,
		Offset:   0,
	})
	if err != nil {
		inventory = []db.Inventory{}
	}

	var items []domain.StructuredItem
	var confidence float64
	var parseErr error

	switch req.InputType {
	case db.InputTypeText:
		items, confidence, parseErr = s.aiClient.ParseText(ctx, req.RawText, inventory)
	case db.InputTypeImage:
		items, confidence, parseErr = s.aiClient.ParseImage(ctx, req.MediaUrl, inventory)
	case db.InputTypePdf:
		items, confidence, parseErr = s.aiClient.ParsePDF(ctx, req.MediaUrl, inventory)
	case db.InputTypeVoice:
		if s.transcriber == nil || s.mediaSvc == nil {
			parseErr = fmt.Errorf("voice transcription not configured")
			break
		}

		audioReader, _, dlErr := s.mediaSvc.Download(ctx, req.MediaUrl)
		if dlErr != nil {
			parseErr = fmt.Errorf("download audio: %w", dlErr)
			break
		}
		defer audioReader.Close()

		// Limit audio to 25MB to prevent memory exhaustion
		const maxAudioSize = 25 * 1024 * 1024
		limitedReader := io.LimitReader(audioReader, maxAudioSize)

		transcript, tErr := s.transcriber.Transcribe(ctx, limitedReader)
		if tErr != nil {
			parseErr = fmt.Errorf("transcribe audio: %w", tErr)
			break
		}

		// Store the transcript
		if err := s.store.Queries.UpdateRequestRawText(ctx, db.UpdateRequestRawTextParams{
			ID:      requestID,
			RawText: transcript,
		}); err != nil {
			slog.Error("failed to store transcript", "id", requestID, "error", err)
		}

		items, confidence, parseErr = s.aiClient.ParseText(ctx, transcript, inventory)
	}

	if parseErr != nil {
		if statusErr := s.store.Queries.SetRequestFailed(ctx, db.SetRequestFailedParams{
			ID:        requestID,
			LastError: parseErr.Error(),
		}); statusErr != nil {
			slog.Error("failed to set request status to failed", "id", requestID, "error", statusErr)
		}
		return nil, fmt.Errorf("AI parse: %w", parseErr)
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("marshal items: %w", err)
	}

	updated, err := s.store.Queries.UpdateRequestStructuredItemsVersioned(ctx, db.UpdateRequestStructuredItemsVersionedParams{
		ID:              requestID,
		StructuredItems: itemsJSON,
		AiConfidence:    fmt.Sprintf("%.4f", confidence),
		Version:         req.Version,
	})
	if err != nil {
		return nil, fmt.Errorf("update items: %w", wrapVersionErr(err))
	}

	return &updated, nil
}

func (s *RequestService) Confirm(ctx context.Context, requestID uuid.UUID, items json.RawMessage) (*db.Request, error) {
	req, err := s.store.Queries.GetRequest(ctx, requestID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if req.Status != db.RequestStatusParsed {
		return nil, domain.ErrInvalidStatus
	}

	// Use a transaction to atomically update items + status with optimistic concurrency
	var result db.Request
	currentVersion := req.Version
	if txErr := s.store.WithTx(ctx, func(qtx *db.Queries) error {
		if items != nil {
			updated, err := qtx.UpdateRequestStructuredItemsVersioned(ctx, db.UpdateRequestStructuredItemsVersionedParams{
				ID:              requestID,
				StructuredItems: items,
				AiConfidence:    req.AiConfidence,
				Version:         currentVersion,
			})
			if err != nil {
				return fmt.Errorf("update items: %w", wrapVersionErr(err))
			}
			currentVersion = updated.Version
		}

		r, err := qtx.UpdateRequestStatusVersioned(ctx, db.UpdateRequestStatusVersionedParams{
			ID:      requestID,
			Status:  db.RequestStatusConfirmed,
			Version: currentVersion,
		})
		if err != nil {
			return fmt.Errorf("confirm: %w", wrapVersionErr(err))
		}
		result = r
		return nil
	}); txErr != nil {
		return nil, txErr
	}

	return &result, nil
}

func (s *RequestService) Send(ctx context.Context, requestID uuid.UUID) (*db.Request, error) {
	req, err := s.store.Queries.GetRequest(ctx, requestID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if req.Status != db.RequestStatusConfirmed {
		return nil, domain.ErrInvalidStatus
	}

	// Use a transaction for status update with optimistic concurrency
	var result db.Request
	if txErr := s.store.WithTx(ctx, func(qtx *db.Queries) error {
		r, err := qtx.UpdateRequestStatusVersioned(ctx, db.UpdateRequestStatusVersionedParams{
			ID:      requestID,
			Status:  db.RequestStatusSent,
			Version: req.Version,
		})
		if err != nil {
			return fmt.Errorf("send: %w", wrapVersionErr(err))
		}
		result = r
		return nil
	}); txErr != nil {
		return nil, txErr
	}

	// Send order confirmation email asynchronously (tracked for graceful shutdown)
	if s.emailClient != nil {
		s.emailWg.Add(1)
		go func() {
			defer s.emailWg.Done()

			emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			dealer, dErr := s.store.Queries.GetDealer(emailCtx, result.DealerID)
			if dErr != nil {
				slog.Warn("failed to get dealer for email", "error", dErr, "request_id", requestID)
				return
			}

			if dealer.ContactEmail != "" {
				var structuredItems []domain.StructuredItem
				if len(result.StructuredItems) > 0 {
					if jErr := json.Unmarshal(result.StructuredItems, &structuredItems); jErr != nil {
						slog.Warn("failed to unmarshal items for email", "error", jErr, "request_id", requestID)
						return
					}
				}

				if eErr := s.emailClient.SendOrderConfirmation(emailCtx, dealer.ContactEmail, dealer.Name, structuredItems); eErr != nil {
					slog.Error("failed to send order confirmation email", "error", eErr, "request_id", requestID)
				} else {
					slog.Info("order confirmation email sent", "request_id", requestID, "to", dealer.ContactEmail)
				}
			}
		}()
	}

	return &result, nil
}
