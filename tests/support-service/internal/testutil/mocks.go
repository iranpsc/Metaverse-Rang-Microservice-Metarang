package testutil

import (
	"context"

	"metarang/support-service/internal/models"
	"metarang/support-service/internal/repository"
)

// MockTicketRepo implements repository.TicketRepository for tests.
type MockTicketRepo struct {
	CreateFunc                  func(ctx context.Context, ticket *models.Ticket) (*models.Ticket, error)
	GetByIDFunc                 func(ctx context.Context, ticketID uint64) (*models.TicketWithRelations, error)
	GetByUserIDFunc             func(ctx context.Context, userID uint64, page, perPage int32, received bool) ([]*models.TicketWithRelations, int, error)
	UpdateFunc                  func(ctx context.Context, ticket *models.Ticket) error
	UpdateStatusFunc            func(ctx context.Context, ticketID uint64, status int32) error
	GetResponsesByTicketIDFunc  func(ctx context.Context, ticketID uint64) ([]models.TicketResponse, error)
	CreateResponseFunc          func(ctx context.Context, response *models.TicketResponse) (*models.TicketResponse, error)
	CheckUserOwnershipFunc      func(ctx context.Context, ticketID, userID uint64) (bool, error)
	GetTicketSenderReceiverFunc func(ctx context.Context, ticketID uint64) (senderID, receiverID uint64, err error)
}

func (m *MockTicketRepo) Create(ctx context.Context, ticket *models.Ticket) (*models.Ticket, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, ticket)
	}
	return ticket, nil
}

func (m *MockTicketRepo) GetByID(ctx context.Context, ticketID uint64) (*models.TicketWithRelations, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, ticketID)
	}
	return nil, nil
}

func (m *MockTicketRepo) GetByUserID(ctx context.Context, userID uint64, page, perPage int32, received bool) ([]*models.TicketWithRelations, int, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID, page, perPage, received)
	}
	return nil, 0, nil
}

func (m *MockTicketRepo) Update(ctx context.Context, ticket *models.Ticket) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, ticket)
	}
	return nil
}

func (m *MockTicketRepo) UpdateStatus(ctx context.Context, ticketID uint64, status int32) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, ticketID, status)
	}
	return nil
}

func (m *MockTicketRepo) GetResponsesByTicketID(ctx context.Context, ticketID uint64) ([]models.TicketResponse, error) {
	if m.GetResponsesByTicketIDFunc != nil {
		return m.GetResponsesByTicketIDFunc(ctx, ticketID)
	}
	return nil, nil
}

func (m *MockTicketRepo) CreateResponse(ctx context.Context, response *models.TicketResponse) (*models.TicketResponse, error) {
	if m.CreateResponseFunc != nil {
		return m.CreateResponseFunc(ctx, response)
	}
	return response, nil
}

func (m *MockTicketRepo) CheckUserOwnership(ctx context.Context, ticketID, userID uint64) (bool, error) {
	if m.CheckUserOwnershipFunc != nil {
		return m.CheckUserOwnershipFunc(ctx, ticketID, userID)
	}
	return false, nil
}

func (m *MockTicketRepo) GetTicketSenderReceiver(ctx context.Context, ticketID uint64) (uint64, uint64, error) {
	if m.GetTicketSenderReceiverFunc != nil {
		return m.GetTicketSenderReceiverFunc(ctx, ticketID)
	}
	return 0, 0, nil
}

var _ repository.TicketRepository = (*MockTicketRepo)(nil)

// MockReportRepo implements repository.ReportRepository for tests.
type MockReportRepo struct {
	CreateFunc      func(ctx context.Context, report *models.Report) (*models.Report, error)
	GetByIDFunc     func(ctx context.Context, reportID uint64) (*models.ReportWithImages, error)
	GetByUserIDFunc func(ctx context.Context, userID uint64, page, perPage int32) ([]*models.Report, int, error)
	CreateImageFunc func(ctx context.Context, reportID uint64, url string) error
}

func (m *MockReportRepo) Create(ctx context.Context, report *models.Report) (*models.Report, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, report)
	}
	return report, nil
}

func (m *MockReportRepo) GetByID(ctx context.Context, reportID uint64) (*models.ReportWithImages, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, reportID)
	}
	return nil, nil
}

func (m *MockReportRepo) GetByUserID(ctx context.Context, userID uint64, page, perPage int32) ([]*models.Report, int, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID, page, perPage)
	}
	return nil, 0, nil
}

func (m *MockReportRepo) CreateImage(ctx context.Context, reportID uint64, url string) error {
	if m.CreateImageFunc != nil {
		return m.CreateImageFunc(ctx, reportID, url)
	}
	return nil
}

var _ repository.ReportRepository = (*MockReportRepo)(nil)

// MockNoteRepo implements repository.NoteRepository for tests.
type MockNoteRepo struct {
	CreateFunc             func(ctx context.Context, note *models.Note) (*models.Note, error)
	GetByIDFunc            func(ctx context.Context, noteID uint64) (*models.Note, error)
	GetByUserIDFunc        func(ctx context.Context, userID uint64) ([]*models.Note, error)
	UpdateFunc             func(ctx context.Context, note *models.Note) error
	DeleteFunc             func(ctx context.Context, noteID uint64) error
	CheckUserOwnershipFunc func(ctx context.Context, noteID, userID uint64) (bool, error)
}

func (m *MockNoteRepo) Create(ctx context.Context, note *models.Note) (*models.Note, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, note)
	}
	return note, nil
}

func (m *MockNoteRepo) GetByID(ctx context.Context, noteID uint64) (*models.Note, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, noteID)
	}
	return nil, nil
}

func (m *MockNoteRepo) GetByUserID(ctx context.Context, userID uint64) ([]*models.Note, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockNoteRepo) Update(ctx context.Context, note *models.Note) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, note)
	}
	return nil
}

func (m *MockNoteRepo) Delete(ctx context.Context, noteID uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, noteID)
	}
	return nil
}

func (m *MockNoteRepo) CheckUserOwnership(ctx context.Context, noteID, userID uint64) (bool, error) {
	if m.CheckUserOwnershipFunc != nil {
		return m.CheckUserOwnershipFunc(ctx, noteID, userID)
	}
	return false, nil
}

var _ repository.NoteRepository = (*MockNoteRepo)(nil)

// MockUserEventRepo implements repository.UserEventRepository for tests.
type MockUserEventRepo struct {
	CreateFunc               func(ctx context.Context, event *models.UserEvent) (*models.UserEvent, error)
	GetByIDFunc              func(ctx context.Context, eventID uint64) (*models.UserEventWithReport, error)
	GetByUserIDFunc          func(ctx context.Context, userID uint64, page, perPage int32) ([]*models.UserEvent, int, error)
	CreateReportFunc         func(ctx context.Context, report *models.UserEventReport) (*models.UserEventReport, error)
	UpdateReportStatusFunc   func(ctx context.Context, reportID uint64, status int32) error
	CloseReportFunc          func(ctx context.Context, reportID uint64) error
	CreateReportResponseFunc func(ctx context.Context, response *models.UserEventReportResponse) (*models.UserEventReportResponse, error)
	GetReportByEventIDFunc   func(ctx context.Context, eventID uint64) (*models.UserEventReport, error)
	GetReportResponsesFunc   func(ctx context.Context, reportID uint64) ([]models.UserEventReportResponse, error)
}

func (m *MockUserEventRepo) Create(ctx context.Context, event *models.UserEvent) (*models.UserEvent, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, event)
	}
	return event, nil
}

func (m *MockUserEventRepo) GetByID(ctx context.Context, eventID uint64) (*models.UserEventWithReport, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, eventID)
	}
	return nil, nil
}

func (m *MockUserEventRepo) GetByUserID(ctx context.Context, userID uint64, page, perPage int32) ([]*models.UserEvent, int, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID, page, perPage)
	}
	return nil, 0, nil
}

func (m *MockUserEventRepo) CreateReport(ctx context.Context, report *models.UserEventReport) (*models.UserEventReport, error) {
	if m.CreateReportFunc != nil {
		return m.CreateReportFunc(ctx, report)
	}
	return report, nil
}

func (m *MockUserEventRepo) UpdateReportStatus(ctx context.Context, reportID uint64, status int32) error {
	if m.UpdateReportStatusFunc != nil {
		return m.UpdateReportStatusFunc(ctx, reportID, status)
	}
	return nil
}

func (m *MockUserEventRepo) CloseReport(ctx context.Context, reportID uint64) error {
	if m.CloseReportFunc != nil {
		return m.CloseReportFunc(ctx, reportID)
	}
	return nil
}

func (m *MockUserEventRepo) CreateReportResponse(ctx context.Context, response *models.UserEventReportResponse) (*models.UserEventReportResponse, error) {
	if m.CreateReportResponseFunc != nil {
		return m.CreateReportResponseFunc(ctx, response)
	}
	response.ID = 1
	return response, nil
}

func (m *MockUserEventRepo) GetReportByEventID(ctx context.Context, eventID uint64) (*models.UserEventReport, error) {
	if m.GetReportByEventIDFunc != nil {
		return m.GetReportByEventIDFunc(ctx, eventID)
	}
	return nil, nil
}

func (m *MockUserEventRepo) GetReportResponses(ctx context.Context, reportID uint64) ([]models.UserEventReportResponse, error) {
	if m.GetReportResponsesFunc != nil {
		return m.GetReportResponsesFunc(ctx, reportID)
	}
	return nil, nil
}

var _ repository.UserEventRepository = (*MockUserEventRepo)(nil)
