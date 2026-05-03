package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"metargb/grpc-gateway/internal/middleware"
	pbAuth "metargb/shared/pb/auth"
	pbCommon "metargb/shared/pb/common"
	pbSupport "metargb/shared/pb/support"
)

type SupportHandler struct {
	ticketClient    pbSupport.TicketServiceClient
	reportClient    pbSupport.ReportServiceClient
	userEventClient pbSupport.UserEventReportServiceClient
	noteClient      pbSupport.NoteServiceClient
	authClient      pbAuth.AuthServiceClient
}

func NewSupportHandler(supportConn, authConn *grpc.ClientConn) *SupportHandler {
	return &SupportHandler{
		ticketClient:    pbSupport.NewTicketServiceClient(supportConn),
		reportClient:    pbSupport.NewReportServiceClient(supportConn),
		userEventClient: pbSupport.NewUserEventReportServiceClient(supportConn),
		noteClient:      pbSupport.NewNoteServiceClient(supportConn),
		authClient:      pbAuth.NewAuthServiceClient(authConn),
	}
}

// Helper function to get authenticated user ID from context (set by auth middleware)
func (h *SupportHandler) getAuthUserID(r *http.Request) (uint64, error) {
	userCtx, err := middleware.GetUserFromRequest(r)
	if err != nil {
		return 0, err
	}
	return userCtx.UserID, nil
}

func splitJalaliDateTime(s string) (date, clock string) {
	parts := strings.SplitN(strings.TrimSpace(s), " ", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return s, s
}

func userEventStatusLabel(ok bool) string {
	if ok {
		return "موفق"
	}
	return "ناموفق"
}

// displayNameFromAuth returns a short display name for ticket responses (Laravel uses user name).
func (h *SupportHandler) displayNameFromAuth(r *http.Request) string {
	userCtx, err := middleware.GetUserFromRequest(r)
	if err != nil || userCtx == nil || userCtx.Email == "" {
		return "User"
	}
	email := userCtx.Email
	if i := strings.Index(email, "@"); i > 0 {
		return email[:i]
	}
	return email
}

// ============================================================================
// Tickets API
// ============================================================================

// ListTickets handles GET /api/tickets
// Query params: page, cursor, recieved (bool)
func (h *SupportHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	// Parse pagination
	page := int32(1)
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.ParseInt(p, 10, 32); err == nil {
			page = int32(parsed)
		}
	}

	perPage := int32(10)
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.ParseInt(pp, 10, 32); err == nil {
			perPage = int32(parsed)
		}
	}

	received := r.URL.Query().Get("recieved") == "true" || r.URL.Query().Get("recieved") == "1"

	grpcReq := &pbSupport.GetTicketsRequest{
		UserId: userID,
		Pagination: &pbCommon.PaginationRequest{
			Page:    page,
			PerPage: perPage,
		},
		Received: received,
	}

	resp, err := h.ticketClient.GetTickets(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	// Convert to Laravel-compatible format
	tickets := make([]map[string]interface{}, 0, len(resp.Tickets))
	for _, ticket := range resp.Tickets {
		dateStr, timeStr := splitJalaliDateTime(ticket.UpdatedAt)
		ticketMap := map[string]interface{}{
			"id":      ticket.Id,
			"title":   ticket.Title,
			"content": ticket.Content,
			"code":    ticket.Code,
			"status":  ticket.Status,
			"date":    dateStr,
			"time":    timeStr,
		}

		if ticket.Sender != nil {
			ticketMap["sender"] = map[string]interface{}{
				"name":          ticket.Sender.Name,
				"code":          ticket.Sender.Code,
				"profile-photo": ticket.Sender.ProfilePhoto,
			}
		}

		if ticket.Receiver != nil {
			ticketMap["reciever"] = map[string]interface{}{
				"name":          ticket.Receiver.Name,
				"code":          ticket.Receiver.Code,
				"profile-photo": ticket.Receiver.ProfilePhoto,
			}
		}

		if ticket.Department != "" {
			ticketMap["department"] = ticket.Department
		}

		if ticket.Attachment != "" {
			ticketMap["attachment"] = ticket.Attachment
		}

		if len(ticket.Responses) > 0 {
			responses := make([]map[string]interface{}, 0, len(ticket.Responses))
			for _, resp := range ticket.Responses {
				responses = append(responses, map[string]interface{}{
					"id":             resp.Id,
					"response":       resp.Response,
					"attachment":     resp.Attachment,
					"responser_name": resp.ResponserName,
					"responser_id":   resp.ResponserId,
					"created_at":     resp.CreatedAt,
				})
			}
			ticketMap["responses"] = responses
		}

		tickets = append(tickets, ticketMap)
	}

	// Simple pagination response (matching Laravel simplePaginate)
	response := map[string]interface{}{
		"data": tickets,
	}
	if len(tickets) == int(perPage) {
		response["next_page_url"] = r.URL.Path + "?page=" + strconv.Itoa(int(page+1))
	}

	writeJSON(w, http.StatusOK, response)
}

// CreateTicket handles POST /api/tickets
func (h *SupportHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	var req struct {
		Title      string  `json:"title"`
		Content    string  `json:"content"`
		Attachment string  `json:"attachment"`
		Reciever   *uint64 `json:"reciever"` // Note: Laravel uses 'reciever' (typo)
		Department string  `json:"department"`
	}

	if err := decodeRequestBody(r, &req); err != nil {
		if err == io.EOF {
			writeError(w, http.StatusBadRequest, "request body is required")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}

	grpcReq := &pbSupport.CreateTicketRequest{
		UserId:     userID,
		Title:      req.Title,
		Content:    req.Content,
		Attachment: req.Attachment,
	}

	if req.Reciever != nil {
		grpcReq.ReceiverId = *req.Reciever
	}
	if req.Department != "" {
		grpcReq.Department = req.Department
	}

	resp, err := h.ticketClient.CreateTicket(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	dateStr, timeStr := splitJalaliDateTime(resp.UpdatedAt)
	ticketMap := map[string]interface{}{
		"id":      resp.Id,
		"title":   resp.Title,
		"content": resp.Content,
		"code":    resp.Code,
		"status":  resp.Status,
		"date":    dateStr,
		"time":    timeStr,
	}

	if resp.Sender != nil {
		ticketMap["sender"] = map[string]interface{}{
			"name":          resp.Sender.Name,
			"code":          resp.Sender.Code,
			"profile-photo": resp.Sender.ProfilePhoto,
		}
	}

	if resp.Receiver != nil {
		ticketMap["reciever"] = map[string]interface{}{
			"name":          resp.Receiver.Name,
			"code":          resp.Receiver.Code,
			"profile-photo": resp.Receiver.ProfilePhoto,
		}
	}

	if resp.Department != "" {
		ticketMap["department"] = resp.Department
	}

	writeJSON(w, http.StatusCreated, ticketMap)
}

// GetTicket handles GET /api/tickets/{ticket}
func (h *SupportHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	ticketIDStr := extractIDFromPath(r.URL.Path, "/api/tickets/")
	if ticketIDStr == "" {
		writeError(w, http.StatusBadRequest, "ticket_id is required")
		return
	}

	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket_id")
		return
	}

	grpcReq := &pbSupport.GetTicketRequest{
		TicketId: ticketID,
		UserId:   userID,
	}

	resp, err := h.ticketClient.GetTicket(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	dateStr, timeStr := splitJalaliDateTime(resp.UpdatedAt)
	ticketMap := map[string]interface{}{
		"id":      resp.Id,
		"title":   resp.Title,
		"content": resp.Content,
		"code":    resp.Code,
		"status":  resp.Status,
		"date":    dateStr,
		"time":    timeStr,
	}

	if resp.Sender != nil {
		ticketMap["sender"] = map[string]interface{}{
			"name":          resp.Sender.Name,
			"code":          resp.Sender.Code,
			"profile-photo": resp.Sender.ProfilePhoto,
		}
	}

	if resp.Receiver != nil {
		ticketMap["reciever"] = map[string]interface{}{
			"name":          resp.Receiver.Name,
			"code":          resp.Receiver.Code,
			"profile-photo": resp.Receiver.ProfilePhoto,
		}
	}

	if resp.Department != "" {
		ticketMap["department"] = resp.Department
	}

	if len(resp.Responses) > 0 {
		responses := make([]map[string]interface{}, 0, len(resp.Responses))
		for _, resp := range resp.Responses {
			responses = append(responses, map[string]interface{}{
				"id":             resp.Id,
				"response":       resp.Response,
				"attachment":     resp.Attachment,
				"responser_name": resp.ResponserName,
				"responser_id":   resp.ResponserId,
				"created_at":     resp.CreatedAt,
			})
		}
		ticketMap["responses"] = responses
	}

	writeJSON(w, http.StatusOK, ticketMap)
}

// UpdateTicket handles PUT/PATCH /api/tickets/{ticket}
func (h *SupportHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	ticketIDStr := extractIDFromPath(r.URL.Path, "/api/tickets/")
	if ticketIDStr == "" {
		writeError(w, http.StatusBadRequest, "ticket_id is required")
		return
	}

	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket_id")
		return
	}

	var req struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		Attachment string `json:"attachment"`
	}

	if err := decodeRequestBody(r, &req); err != nil {
		if err == io.EOF {
			writeError(w, http.StatusBadRequest, "request body is required")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}

	grpcReq := &pbSupport.UpdateTicketRequest{
		TicketId:   ticketID,
		UserId:     userID,
		Title:      req.Title,
		Content:    req.Content,
		Attachment: req.Attachment,
	}

	resp, err := h.ticketClient.UpdateTicket(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	dateStr, timeStr := splitJalaliDateTime(resp.UpdatedAt)
	ticketMap := map[string]interface{}{
		"id":      resp.Id,
		"title":   resp.Title,
		"content": resp.Content,
		"code":    resp.Code,
		"status":  resp.Status,
		"date":    dateStr,
		"time":    timeStr,
	}

	if resp.Sender != nil {
		ticketMap["sender"] = map[string]interface{}{
			"name":          resp.Sender.Name,
			"code":          resp.Sender.Code,
			"profile-photo": resp.Sender.ProfilePhoto,
		}
	}

	writeJSON(w, http.StatusOK, ticketMap)
}

// AddTicketResponse handles POST /api/tickets/response/{ticket}
func (h *SupportHandler) AddTicketResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	ticketIDStr := extractIDFromPath(r.URL.Path, "/api/tickets/response/")
	if ticketIDStr == "" {
		writeError(w, http.StatusBadRequest, "ticket_id is required")
		return
	}

	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket_id")
		return
	}

	var req struct {
		Response   string `json:"response"`
		Attachment string `json:"attachment"`
	}

	if err := decodeRequestBody(r, &req); err != nil {
		if err == io.EOF {
			writeError(w, http.StatusBadRequest, "request body is required")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}

	grpcReq := &pbSupport.AddResponseRequest{
		TicketId:   ticketID,
		UserId:     userID,
		Response:   req.Response,
		Attachment: req.Attachment,
		UserName:   h.displayNameFromAuth(r),
	}

	resp, err := h.ticketClient.AddResponse(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	dateStr, timeStr := splitJalaliDateTime(resp.UpdatedAt)
	ticketMap := map[string]interface{}{
		"id":      resp.Id,
		"title":   resp.Title,
		"content": resp.Content,
		"code":    resp.Code,
		"status":  resp.Status,
		"date":    dateStr,
		"time":    timeStr,
	}

	if len(resp.Responses) > 0 {
		responses := make([]map[string]interface{}, 0, len(resp.Responses))
		for _, resp := range resp.Responses {
			responses = append(responses, map[string]interface{}{
				"id":             resp.Id,
				"response":       resp.Response,
				"attachment":     resp.Attachment,
				"responser_name": resp.ResponserName,
				"responser_id":   resp.ResponserId,
				"created_at":     resp.CreatedAt,
			})
		}
		ticketMap["responses"] = responses
	}

	writeJSON(w, http.StatusOK, ticketMap)
}

// CloseTicket handles GET /api/tickets/close/{ticket}
func (h *SupportHandler) CloseTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	ticketIDStr := extractIDFromPath(r.URL.Path, "/api/tickets/close/")
	if ticketIDStr == "" {
		writeError(w, http.StatusBadRequest, "ticket_id is required")
		return
	}

	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket_id")
		return
	}

	grpcReq := &pbSupport.CloseTicketRequest{
		TicketId: ticketID,
		UserId:   userID,
	}

	resp, err := h.ticketClient.CloseTicket(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	dateStr, timeStr := splitJalaliDateTime(resp.UpdatedAt)
	ticketMap := map[string]interface{}{
		"id":      resp.Id,
		"title":   resp.Title,
		"content": resp.Content,
		"code":    resp.Code,
		"status":  resp.Status,
		"date":    dateStr,
		"time":    timeStr,
	}

	writeJSON(w, http.StatusOK, ticketMap)
}

// ============================================================================
// Reports API
// ============================================================================

// ListReports handles GET /api/reports
func (h *SupportHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	page := int32(1)
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.ParseInt(p, 10, 32); err == nil {
			page = int32(parsed)
		}
	}

	perPage := int32(10)
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.ParseInt(pp, 10, 32); err == nil {
			perPage = int32(parsed)
		}
	}

	grpcReq := &pbSupport.GetReportsRequest{
		UserId: userID,
		Pagination: &pbCommon.PaginationRequest{
			Page:    page,
			PerPage: perPage,
		},
	}

	resp, err := h.reportClient.GetReports(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	reports := make([]map[string]interface{}, 0, len(resp.Reports))
	for _, report := range resp.Reports {
		reportMap := map[string]interface{}{
			"id":       report.Id,
			"title":    report.Reason,         // Mapping reason to title
			"subject":  report.ReportableType, // Mapping reportable_type to subject
			"content":  report.Description,
			"datetime": report.CreatedAt,
		}
		reports = append(reports, reportMap)
	}

	response := map[string]interface{}{
		"data": reports,
	}
	if len(reports) == int(perPage) {
		response["next_page_url"] = r.URL.Path + "?page=" + strconv.Itoa(int(page+1))
	}

	writeJSON(w, http.StatusOK, response)
}

// CreateReport handles POST /api/reports
func (h *SupportHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	var req struct {
		Subject     string   `json:"subject"`
		Title       string   `json:"title"`
		Content     string   `json:"content"`
		URL         string   `json:"url"`
		Attachments []string `json:"attachments"`
	}

	if err := decodeRequestBody(r, &req); err != nil {
		if err == io.EOF {
			writeError(w, http.StatusBadRequest, "request body is required")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}

	grpcReq := &pbSupport.CreateReportRequest{
		UserId:         userID,
		ReportableType: req.Subject,
		Reason:         req.Title,
		Description:    req.Content,
		Url:            req.URL,
		ImageUrls:      req.Attachments,
	}

	resp, err := h.reportClient.CreateReport(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	reportMap := map[string]interface{}{
		"id":       strconv.FormatUint(resp.Id, 10),
		"title":    resp.Reason,
		"subject":  resp.ReportableType,
		"content":  resp.Description,
		"datetime": resp.CreatedAt,
	}

	if resp.Url != "" {
		reportMap["url"] = resp.Url
	}
	if len(resp.AttachmentUrls) > 0 {
		reportMap["attachments"] = resp.AttachmentUrls
	}

	writeJSON(w, http.StatusCreated, reportMap)
}

// GetReport handles GET /api/reports/{report}
func (h *SupportHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	reportIDStr := extractIDFromPath(r.URL.Path, "/api/reports/")
	if reportIDStr == "" {
		writeError(w, http.StatusBadRequest, "report_id is required")
		return
	}

	reportID, err := strconv.ParseUint(reportIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid report_id")
		return
	}

	grpcReq := &pbSupport.GetReportRequest{
		ReportId: reportID,
		UserId:   userID,
	}

	resp, err := h.reportClient.GetReport(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	reportMap := map[string]interface{}{
		"id":       strconv.FormatUint(resp.Id, 10),
		"title":    resp.Reason,
		"subject":  resp.ReportableType,
		"content":  resp.Description,
		"datetime": resp.CreatedAt,
	}

	if resp.Url != "" {
		reportMap["url"] = resp.Url
	}
	if len(resp.AttachmentUrls) > 0 {
		reportMap["attachments"] = resp.AttachmentUrls
	}

	writeJSON(w, http.StatusOK, reportMap)
}

// ============================================================================
// Notes API
// ============================================================================

// ListNotes handles GET /api/notes
func (h *SupportHandler) ListNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	grpcReq := &pbSupport.GetNotesRequest{
		UserId: userID,
	}

	resp, err := h.noteClient.GetNotes(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	notes := make([]map[string]interface{}, 0, len(resp.Notes))
	for _, note := range resp.Notes {
		noteMap := map[string]interface{}{
			"id":      note.Id,
			"title":   note.Title,
			"content": note.Content,
			"date":    note.Date,
			"time":    note.Time,
		}
		if len(note.Attachments) > 0 {
			noteMap["attachments"] = note.Attachments
		}
		notes = append(notes, noteMap)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": notes})
}

// CreateNote handles POST /api/notes
func (h *SupportHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	var req struct {
		Title       string   `json:"title"`
		Content     string   `json:"content"`
		Attachments []string `json:"attachments"`
	}

	if err := decodeRequestBody(r, &req); err != nil {
		if err == io.EOF {
			writeError(w, http.StatusBadRequest, "request body is required")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}

	grpcReq := &pbSupport.CreateNoteRequest{
		UserId:      userID,
		Title:       req.Title,
		Content:     req.Content,
		Attachments: req.Attachments,
	}

	resp, err := h.noteClient.CreateNote(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	noteMap := map[string]interface{}{
		"id":      resp.Id,
		"title":   resp.Title,
		"content": resp.Content,
		"date":    resp.Date,
		"time":    resp.Time,
	}
	if len(resp.Attachments) > 0 {
		noteMap["attachments"] = resp.Attachments
	}

	writeJSON(w, http.StatusCreated, noteMap)
}

// GetNote handles GET /api/notes/{note}
func (h *SupportHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	noteIDStr := extractIDFromPath(r.URL.Path, "/api/notes/")
	if noteIDStr == "" {
		writeError(w, http.StatusBadRequest, "note_id is required")
		return
	}

	noteID, err := strconv.ParseUint(noteIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note_id")
		return
	}

	grpcReq := &pbSupport.GetNoteRequest{
		NoteId: noteID,
		UserId: userID,
	}

	resp, err := h.noteClient.GetNote(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	noteMap := map[string]interface{}{
		"id":      resp.Id,
		"title":   resp.Title,
		"content": resp.Content,
		"date":    resp.Date,
		"time":    resp.Time,
	}
	if len(resp.Attachments) > 0 {
		noteMap["attachments"] = resp.Attachments
	}

	writeJSON(w, http.StatusOK, noteMap)
}

// UpdateNote handles PUT/PATCH /api/notes/{note}
func (h *SupportHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	noteIDStr := extractIDFromPath(r.URL.Path, "/api/notes/")
	if noteIDStr == "" {
		writeError(w, http.StatusBadRequest, "note_id is required")
		return
	}

	noteID, err := strconv.ParseUint(noteIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note_id")
		return
	}

	var req struct {
		Title       string   `json:"title"`
		Content     string   `json:"content"`
		Attachments []string `json:"attachments"`
	}

	if err := decodeRequestBody(r, &req); err != nil {
		if err == io.EOF {
			writeError(w, http.StatusBadRequest, "request body is required")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}

	grpcReq := &pbSupport.UpdateNoteRequest{
		NoteId:      noteID,
		UserId:      userID,
		Title:       req.Title,
		Content:     req.Content,
		Attachments: req.Attachments,
	}

	resp, err := h.noteClient.UpdateNote(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	noteMap := map[string]interface{}{
		"id":      resp.Id,
		"title":   resp.Title,
		"content": resp.Content,
		"date":    resp.Date,
		"time":    resp.Time,
	}
	if len(resp.Attachments) > 0 {
		noteMap["attachments"] = resp.Attachments
	}

	writeJSON(w, http.StatusOK, noteMap)
}

// DeleteNote handles DELETE /api/notes/{note}
func (h *SupportHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	noteIDStr := extractIDFromPath(r.URL.Path, "/api/notes/")
	if noteIDStr == "" {
		writeError(w, http.StatusBadRequest, "note_id is required")
		return
	}

	noteID, err := strconv.ParseUint(noteIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note_id")
		return
	}

	grpcReq := &pbSupport.DeleteNoteRequest{
		NoteId: noteID,
		UserId: userID,
	}

	_, err = h.noteClient.DeleteNote(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// User events (Laravel UserEventsController → support-service)
// ============================================================================

// ListUserEvents handles GET /api/events
func (h *SupportHandler) ListUserEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	page := int32(1)
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.ParseInt(p, 10, 32); err == nil && parsed > 0 {
			page = int32(parsed)
		}
	}
	perPage := int32(10)
	grpcReq := &pbSupport.GetUserEventsRequest{
		UserId: userID,
		Pagination: &pbCommon.PaginationRequest{
			Page:    page,
			PerPage: perPage,
		},
	}
	resp, err := h.userEventClient.GetUserEvents(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	data := make([]map[string]interface{}, 0, len(resp.Events))
	for _, ev := range resp.Events {
		_, timeStr := splitJalaliDateTime(ev.CreatedAt)
		data = append(data, map[string]interface{}{
			"id":     ev.Id,
			"event":  ev.Title,
			"ip":     ev.Ip,
			"device": ev.Device,
			"status": userEventStatusLabel(ev.StatusOk),
			"date":   ev.EventDate,
			"time":   timeStr,
		})
	}
	out := map[string]interface{}{
		"data": data,
		"links": map[string]interface{}{
			"next": nil,
			"prev": nil,
		},
		"meta": map[string]interface{}{
			"current_page": page,
		},
	}
	if len(data) == int(perPage) {
		out["links"] = map[string]interface{}{
			"next": r.URL.Path + "?page=" + strconv.Itoa(int(page+1)),
			"prev": nil,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// GetUserEvent handles GET /api/events/{id}
func (h *SupportHandler) GetUserEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	eventIDStr := extractIDFromPath(r.URL.Path, "/api/events/")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "event_id is required")
		return
	}
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event_id")
		return
	}
	grpcReq := &pbSupport.GetUserEventRequest{
		EventId: eventID,
		UserId:  userID,
	}
	resp, err := h.userEventClient.GetUserEvent(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	_, timeStr := splitJalaliDateTime(resp.CreatedAt)
	out := map[string]interface{}{
		"id":     resp.Id,
		"event":  resp.Title,
		"ip":     resp.Ip,
		"device": resp.Device,
		"status": userEventStatusLabel(resp.StatusOk),
		"date":   resp.EventDate,
		"time":   timeStr,
	}
	if resp.Report != nil {
		rp := resp.Report
		responses := make([]map[string]interface{}, 0, len(rp.Responses))
		for _, rr := range rp.Responses {
			responses = append(responses, map[string]interface{}{
				"id":             rr.Id,
				"responser_name": rr.ResponserName,
				"response":       rr.Response,
				"date":           rr.Date,
				"time":           rr.Time,
			})
		}
		out["report"] = map[string]interface{}{
			"id":                 rp.Id,
			"suspecious_citizen": rp.SuspiciousCitizen,
			"event_description":  rp.EventDescription,
			"status":             rp.Status,
			"closed":             rp.Closed,
			"date":               rp.Date,
			"time":               rp.Time,
			"responses":          responses,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// ReportUserEvent handles POST /api/events/report/{id}
func (h *SupportHandler) ReportUserEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	eventIDStr := extractIDFromPath(r.URL.Path, "/api/events/report/")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "event_id is required")
		return
	}
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event_id")
		return
	}
	var req struct {
		SuspeciousCitizen string `json:"suspecious_citizen,omitempty"`
		EventDescription  string `json:"event_description"`
	}
	if err := decodeRequestBody(r, &req); err != nil {
		if err == io.EOF {
			writeError(w, http.StatusBadRequest, "request body is required")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}
	grpcReq := &pbSupport.ReportUserEventRequest{
		EventId:           eventID,
		ReporterId:        userID,
		SuspiciousCitizen: req.SuspeciousCitizen,
		EventDescription:  req.EventDescription,
	}
	resp, err := h.userEventClient.ReportUserEvent(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	dateStr, timeStr := splitJalaliDateTime(resp.CreatedAt)
	out := map[string]interface{}{
		"id":                 resp.Id,
		"suspecious_citizen": resp.SuspiciousCitizen,
		"event_description":  resp.EventDescription,
		"status":             0,
		"closed":             false,
		"date":               dateStr,
		"time":               timeStr,
		"responses":          []interface{}{},
	}
	writeJSON(w, http.StatusCreated, out)
}

// SendUserEventReportResponse handles POST /api/events/report/response/{id}
func (h *SupportHandler) SendUserEventReportResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	eventIDStr := extractIDFromPath(r.URL.Path, "/api/events/report/response/")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "event_id is required")
		return
	}
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event_id")
		return
	}
	var req struct {
		Response string `json:"response"`
	}
	if err := decodeRequestBody(r, &req); err != nil {
		if err == io.EOF {
			writeError(w, http.StatusBadRequest, "request body is required")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}
	grpcReq := &pbSupport.SendEventReportResponseRequest{
		EventId:       eventID,
		ResponderId:   userID,
		Response:      req.Response,
		ResponderName: h.displayNameFromAuth(r),
	}
	resp, err := h.userEventClient.SendEventReportResponse(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             resp.Id,
		"responser_name": resp.ResponserName,
		"response":       resp.Response,
		"date":           resp.Date,
		"time":           resp.Time,
	})
}

// CloseUserEventReport handles POST /api/events/report/close/{id}
func (h *SupportHandler) CloseUserEventReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, err := h.getAuthUserID(r)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	eventIDStr := extractIDFromPath(r.URL.Path, "/api/events/report/close/")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "event_id is required")
		return
	}
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event_id")
		return
	}
	grpcReq := &pbSupport.CloseUserEventReportRequest{
		EventId: eventID,
		UserId:  userID,
	}
	_, err = h.userEventClient.CloseUserEventReport(r.Context(), grpcReq)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
