// Package handler provides HTTP handlers for the gRPC gateway service.
package handler

import (
	"net/http"

	"google.golang.org/grpc"

	"metarang/grpc-gateway/internal/middleware"
	pb "metarang/shared/pb/auth"
	levelspb "metarang/shared/pb/levels"
)

type AuthHandler struct {
	authClient              pb.AuthServiceClient
	userClient              pb.UserServiceClient
	kycClient               pb.KYCServiceClient
	citizenClient           pb.CitizenServiceClient
	personalInfoClient      pb.PersonalInfoServiceClient
	profileLimitationClient pb.ProfileLimitationServiceClient
	profilePhotoClient      pb.ProfilePhotoServiceClient
	settingsClient          pb.SettingsServiceClient
	userEventsClient        pb.UserEventsServiceClient
	searchClient            pb.SearchServiceClient
	levelClient             levelspb.LevelServiceClient
	locale                  string
}

func NewAuthHandler(conn *grpc.ClientConn, levelConn *grpc.ClientConn, locale string) *AuthHandler {
	h := &AuthHandler{
		authClient:              pb.NewAuthServiceClient(conn),
		userClient:              pb.NewUserServiceClient(conn),
		kycClient:               pb.NewKYCServiceClient(conn),
		citizenClient:           pb.NewCitizenServiceClient(conn),
		personalInfoClient:      pb.NewPersonalInfoServiceClient(conn),
		profileLimitationClient: pb.NewProfileLimitationServiceClient(conn),
		profilePhotoClient:      pb.NewProfilePhotoServiceClient(conn),
		settingsClient:          pb.NewSettingsServiceClient(conn),
		userEventsClient:        pb.NewUserEventsServiceClient(conn),
		searchClient:            pb.NewSearchServiceClient(conn),
		locale:                  locale,
	}
	if levelConn != nil {
		h.levelClient = levelspb.NewLevelServiceClient(levelConn)
	}
	return h
}

// writeGRPCErrorLocale writes gRPC errors using the handler's locale.
func (h *AuthHandler) writeGRPCErrorLocale(w http.ResponseWriter, err error) {
	writeGRPCErrorWithLocale(w, err, h.locale)
}

func (h *AuthHandler) RequireVerifiedEmail(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := middleware.GetUserFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		user, err := h.userClient.GetUser(r.Context(), &pb.GetUserRequest{UserId: userCtx.UserID})
		if err != nil {
			h.writeGRPCErrorLocale(w, err)
			return
		}
		if user.EmailVerifiedAt == nil {
			writeError(w, http.StatusForbidden, "Your email address is not verified.")
			return
		}

		next.ServeHTTP(w, r)
	})
}
