package handler

import (
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "metarang/shared/pb/auth"
	commercialpb "metarang/shared/pb/commercial"
)

var citizenWalletAssetSet = map[string]struct{}{
	"psc": {}, "irr": {}, "red": {}, "blue": {}, "yellow": {}, "satisfaction": {}, "effect": {},
}

// CitizenWalletHandler serves public citizen wallet history HTTP endpoints.
type CitizenWalletHandler struct {
	citizenClient       authpb.CitizenServiceClient
	walletHistoryClient commercialpb.WalletHistoryServiceClient
	locale              string
}

func NewCitizenWalletHandler(authConn, commercialConn *grpc.ClientConn, locale string) *CitizenWalletHandler {
	return &CitizenWalletHandler{
		citizenClient:       authpb.NewCitizenServiceClient(authConn),
		walletHistoryClient: commercialpb.NewWalletHistoryServiceClient(commercialConn),
		locale:              locale,
	}
}

// Handle dispatches /api/citizen/{code}/wallet/history[/summary|/chart].
func (h *CitizenWalletHandler) Handle(w http.ResponseWriter, r *http.Request, code string, rest []string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if code == "" {
		writeError(w, http.StatusBadRequest, "citizen code is required")
		return
	}

	switch {
	case len(rest) >= 1 && rest[0] == "summary":
		h.handleSummary(w, r, code)
	case len(rest) >= 1 && rest[0] == "chart":
		h.handleChart(w, r, code)
	default:
		writeError(w, http.StatusNotFound, "invalid citizen wallet history endpoint")
	}
}

func (h *CitizenWalletHandler) handleSummary(w http.ResponseWriter, r *http.Request, code string) {
	userID, privacy, ok := h.resolveCitizen(w, r, code)
	if !ok {
		return
	}
	period, ok := resolveWalletHistoryPeriod(w, r)
	if !ok {
		return
	}
	assets, ok := parseWalletAssetsQuery(w, r.URL.Query())
	if !ok {
		return
	}

	resp, err := h.walletHistoryClient.GetWalletHistorySummary(r.Context(), &commercialpb.GetWalletHistorySummaryRequest{
		UserId:  userID,
		Period:  period,
		Assets:  assets,
		Privacy: privacy,
	})
	if err != nil {
		writeGRPCErrorWithLocale(w, err, h.locale)
		return
	}

	data := make([]map[string]interface{}, 0, len(resp.Data))
	for _, card := range resp.Data {
		if card.PrivacyRestricted {
			data = append(data, map[string]interface{}{
				"asset":              card.Asset,
				"privacy_restricted": true,
			})
			continue
		}
		data = append(data, map[string]interface{}{
			"asset":              card.Asset,
			"current_balance":    card.CurrentBalance,
			"period_income":      card.PeriodIncome,
			"period_spending":    card.PeriodSpending,
			"growth_percent":     card.GrowthPercent,
			"direction":          card.Direction,
			"privacy_restricted": false,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": data})
}

func (h *CitizenWalletHandler) handleChart(w http.ResponseWriter, r *http.Request, code string) {
	userID, privacy, ok := h.resolveCitizen(w, r, code)
	if !ok {
		return
	}
	period, ok := resolveWalletHistoryPeriod(w, r)
	if !ok {
		return
	}
	assets, ok := parseWalletAssetsQuery(w, r.URL.Query())
	if !ok {
		return
	}

	resp, err := h.walletHistoryClient.GetWalletHistoryChart(r.Context(), &commercialpb.GetWalletHistoryChartRequest{
		UserId:  userID,
		Period:  period,
		Assets:  assets,
		Privacy: privacy,
	})
	if err != nil {
		writeGRPCErrorWithLocale(w, err, h.locale)
		return
	}

	data := make(map[string]interface{}, len(resp.Data))
	for asset, series := range resp.Data {
		data[asset] = map[string]interface{}{
			"income":   chartPointsJSON(series.Income),
			"spending": chartPointsJSON(series.Spending),
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": data})
}

func (h *CitizenWalletHandler) resolveCitizen(w http.ResponseWriter, r *http.Request, code string) (uint64, map[string]int32, bool) {
	info, err := h.citizenClient.GetCitizenUserInfo(r.Context(), &authpb.GetCitizenUserInfoRequest{Code: code})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			writeError(w, http.StatusNotFound, "citizen not found")
			return 0, nil, false
		}
		writeGRPCErrorWithLocale(w, err, h.locale)
		return 0, nil, false
	}
	return info.UserId, info.Privacy, true
}

func resolveWalletHistoryPeriod(w http.ResponseWriter, r *http.Request) (string, bool) {
	period := strings.TrimSpace(r.URL.Query().Get("period"))
	switch period {
	case "daily", "weekly", "monthly", "yearly":
		return period, true
	case "":
		writeError(w, http.StatusUnprocessableEntity, "The period field is required.")
		return "", false
	default:
		writeError(w, http.StatusUnprocessableEntity, "The selected period is invalid.")
		return "", false
	}
}

func parseWalletAssetsQuery(w http.ResponseWriter, query url.Values) ([]string, bool) {
	raw := []string{}
	if indexed := parseIndexedQueryArray(query, "assets"); len(indexed) > 0 {
		raw = indexed
	} else if vals, ok := query["assets[]"]; ok {
		raw = vals
	} else if vals, ok := query["assets"]; ok {
		raw = vals
	} else if asset := strings.TrimSpace(query.Get("asset")); asset != "" {
		raw = []string{asset}
	}

	if len(raw) == 0 {
		return nil, true
	}

	out := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, asset := range raw {
		asset = strings.ToLower(strings.TrimSpace(asset))
		if asset == "" {
			continue
		}
		if _, ok := citizenWalletAssetSet[asset]; !ok {
			writeError(w, http.StatusUnprocessableEntity, "The selected assets is invalid.")
			return nil, false
		}
		if _, dup := seen[asset]; dup {
			continue
		}
		seen[asset] = struct{}{}
		out = append(out, asset)
	}
	return out, true
}

func chartPointsJSON(points []*commercialpb.WalletChartPoint) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(points))
	for _, p := range points {
		out = append(out, map[string]interface{}{
			"label":  p.Label,
			"amount": p.Amount,
		})
	}
	return out
}
