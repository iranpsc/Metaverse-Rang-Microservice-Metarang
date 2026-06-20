package handler

import (
	"context"
	"metargb/support-service/internal/models"
	"metargb/support-service/internal/service"
	"metargb/support-service/internal/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pbCommon "metargb/shared/pb/common"
	pb "metargb/shared/pb/support"
)

type ReportHandler struct {
	pb.UnimplementedReportServiceServer
	reportService service.ReportService
}

func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

func RegisterReportHandler(grpcServer *grpc.Server, reportService service.ReportService) {
	handler := NewReportHandler(reportService)
	pb.RegisterReportServiceServer(grpcServer, handler)
}

func (h *ReportHandler) CreateReport(ctx context.Context, req *pb.CreateReportRequest) (*pb.ReportResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Reason == "" {
		return nil, status.Error(codes.InvalidArgument, "reason is required")
	}

	report, err := h.reportService.CreateReport(ctx, req.UserId, req.ReportableType, req.Reason, req.Description, req.Url, req.ImagePaths)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create report: %v", err)
	}

	return convertReportWithImagesToProto(report), nil
}

func (h *ReportHandler) GetReports(ctx context.Context, req *pb.GetReportsRequest) (*pb.ReportsResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	page := int32(1)
	perPage := int32(10)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PerPage > 0 {
			perPage = req.Pagination.PerPage
		}
	}

	reports, total, err := h.reportService.GetReports(ctx, req.UserId, page, perPage)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get reports: %v", err)
	}

	response := &pb.ReportsResponse{
		Reports: make([]*pb.ReportResponse, len(reports)),
		Pagination: &pbCommon.PaginationMeta{
			CurrentPage: page,
			PerPage:     perPage,
			Total:       int32(total),
			LastPage:    int32((total + int(perPage) - 1) / int(perPage)),
		},
	}

	for i, report := range reports {
		response.Reports[i] = convertReportToProto(report)
	}

	return response, nil
}

func (h *ReportHandler) GetReport(ctx context.Context, req *pb.GetReportRequest) (*pb.ReportResponse, error) {
	if req.ReportId == 0 {
		return nil, status.Error(codes.InvalidArgument, "report_id is required")
	}

	report, err := h.reportService.GetReport(ctx, req.ReportId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get report: %v", err)
	}

	if report == nil {
		return nil, status.Error(codes.NotFound, "report not found")
	}

	return convertReportWithImagesToProto(report), nil
}

// Helper function to convert report model to proto response
func convertReportToProto(report *models.Report) *pb.ReportResponse {
	return &pb.ReportResponse{
		Id:             report.ID,
		UserId:         report.UserID,
		ReportableType: report.Subject,
		ReportableId:   0,
		Reason:         report.Title,
		Description:    report.Content,
		Url:            report.URL,
		CreatedAt:      utils.FormatJalaliDateTime(report.CreatedAt),
	}
}

func convertReportWithImagesToProto(report *models.ReportWithImages) *pb.ReportResponse {
	resp := convertReportToProto(&report.Report)
	if len(report.Images) > 0 {
		resp.ImagePaths = make([]string, len(report.Images))
		for i, img := range report.Images {
			resp.ImagePaths[i] = img.URL
		}
	}
	return resp
}
