package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"
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

	if req.Status != db.RequestStatusPending {
		return nil, domain.ErrInvalidStatus
	}

	// Set to processing
	req, err = s.store.Queries.UpdateRequestStatus(ctx, db.UpdateRequestStatusParams{
		ID:     requestID,
		Status: db.RequestStatusProcessing,
	})
	if err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	// Get dealer inventory for SKU matching
	inventory, err := s.store.Queries.ListInventory(ctx, db.ListInventoryParams{
		DealerID: req.DealerID,
		Limit:    10000,
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

		transcript, tErr := s.transcriber.Transcribe(ctx, audioReader)
		if tErr != nil {
			parseErr = fmt.Errorf("transcribe audio: %w", tErr)
			break
		}

		// Store the transcript
		s.store.Queries.UpdateRequestRawText(ctx, db.UpdateRequestRawTextParams{
			ID:      requestID,
			RawText: transcript,
		})

		items, confidence, parseErr = s.aiClient.ParseText(ctx, transcript, inventory)
	}

	if parseErr != nil {
		s.store.Queries.UpdateRequestStatus(ctx, db.UpdateRequestStatusParams{
			ID:     requestID,
			Status: db.RequestStatusFailed,
		})
		return nil, fmt.Errorf("AI parse: %w", parseErr)
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("marshal items: %w", err)
	}

	req, err = s.store.Queries.UpdateRequestStructuredItems(ctx, db.UpdateRequestStructuredItemsParams{
		ID:              requestID,
		StructuredItems: itemsJSON,
		AiConfidence:    fmt.Sprintf("%.4f", confidence),
	})
	if err != nil {
		return nil, fmt.Errorf("update items: %w", err)
	}

	return &req, nil
}

func (s *RequestService) Confirm(ctx context.Context, requestID uuid.UUID, items json.RawMessage) (*db.Request, error) {
	req, err := s.store.Queries.GetRequest(ctx, requestID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if req.Status != db.RequestStatusParsed {
		return nil, domain.ErrInvalidStatus
	}

	if items != nil {
		req, err = s.store.Queries.UpdateRequestStructuredItems(ctx, db.UpdateRequestStructuredItemsParams{
			ID:              requestID,
			StructuredItems: items,
			AiConfidence:    req.AiConfidence,
		})
		if err != nil {
			return nil, fmt.Errorf("update items: %w", err)
		}
	}

	req, err = s.store.Queries.UpdateRequestStatus(ctx, db.UpdateRequestStatusParams{
		ID:     requestID,
		Status: db.RequestStatusConfirmed,
	})
	if err != nil {
		return nil, fmt.Errorf("confirm: %w", err)
	}

	return &req, nil
}

func (s *RequestService) Send(ctx context.Context, requestID uuid.UUID) (*db.Request, error) {
	req, err := s.store.Queries.GetRequest(ctx, requestID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if req.Status != db.RequestStatusConfirmed {
		return nil, domain.ErrInvalidStatus
	}

	req, err = s.store.Queries.UpdateRequestStatus(ctx, db.UpdateRequestStatusParams{
		ID:     requestID,
		Status: db.RequestStatusSent,
	})
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	// Send order confirmation email if configured
	if s.emailClient != nil {
		dealer, dErr := s.store.Queries.GetDealer(ctx, req.DealerID)
		if dErr != nil {
			slog.Warn("failed to get dealer for email", "error", dErr)
			return &req, nil
		}

		if dealer.ContactEmail != "" {
			var structuredItems []domain.StructuredItem
			if len(req.StructuredItems) > 0 {
				if jErr := json.Unmarshal(req.StructuredItems, &structuredItems); jErr != nil {
					slog.Warn("failed to unmarshal items for email", "error", jErr)
					return &req, nil
				}
			}

			if eErr := s.emailClient.SendOrderConfirmation(ctx, dealer.ContactEmail, dealer.Name, structuredItems); eErr != nil {
				slog.Warn("failed to send order confirmation email", "error", eErr)
			}
		}
	}

	return &req, nil
}
