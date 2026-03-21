package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"community-aid-api/internal/helpers"
	"community-aid-api/internal/models"
	"community-aid-api/internal/services"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userSvc     *services.UserService
	requestSvc  *services.RequestService
	offerSvc    *services.OfferService
	donationSvc *services.DonationService
	db          *sql.DB
}

func NewAdminHandler(
	userSvc *services.UserService,
	requestSvc *services.RequestService,
	offerSvc *services.OfferService,
	donationSvc *services.DonationService,
	db *sql.DB,
) *AdminHandler {
	return &AdminHandler{
		userSvc:     userSvc,
		requestSvc:  requestSvc,
		offerSvc:    offerSvc,
		donationSvc: donationSvc,
		db:          db,
	}
}

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	tx, err := h.db.BeginTx(c.Request.Context(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		log.Printf("ERROR GetDashboardStats begin tx: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}
	defer tx.Rollback()

	var stats models.DashboardStats
	err = tx.QueryRowContext(c.Request.Context(), `
		SELECT
			(SELECT COUNT(*) FROM users WHERE role = 'community_member')                          AS total_users,
			(SELECT COUNT(*) FROM users WHERE role = 'community_member' AND is_active = true)     AS active_users,
			(SELECT COUNT(*) FROM emergency_requests)                                              AS total_requests,
			(SELECT COUNT(*) FROM emergency_requests WHERE status = 'pending')                    AS pending_requests,
			(SELECT COUNT(*) FROM emergency_requests WHERE status = 'approved')                   AS approved_requests,
			(SELECT COUNT(*) FROM emergency_requests WHERE status = 'rejected')                   AS rejected_requests,
			(SELECT COUNT(*) FROM emergency_requests WHERE status = 'closed')                     AS closed_requests,
			(SELECT COUNT(*) FROM offers)                                                          AS total_offers,
			(SELECT COUNT(*) FROM offers WHERE status = 'pending')                                AS pending_offers,
			(SELECT COUNT(*) FROM offers WHERE status = 'accepted')                               AS accepted_offers,
			(SELECT COUNT(*) FROM offers WHERE status = 'fulfilled')                              AS fulfilled_offers,
			(SELECT COUNT(*) FROM donations)                                                       AS total_donations,
			(SELECT COALESCE(SUM(amount), 0) FROM donations)                                      AS total_donation_amount
	`).Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.TotalRequests,
		&stats.PendingRequests,
		&stats.ApprovedRequests,
		&stats.RejectedRequests,
		&stats.ClosedRequests,
		&stats.TotalOffers,
		&stats.PendingOffers,
		&stats.AcceptedOffers,
		&stats.FulfilledOffers,
		&stats.TotalDonations,
		&stats.TotalDonationAmount,
	)
	if err != nil {
		log.Printf("ERROR GetDashboardStats query: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("ERROR GetDashboardStats commit: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, stats)
}

func (h *AdminHandler) GetAllRequestsAdmin(c *gin.Context) {
	filters := models.RequestFilters{}
	if v := c.Query("type"); v != "" {
		filters.Type = &v
	}
	if v := c.Query("status"); v != "" {
		filters.Status = &v
	}
	if v := c.Query("location_name"); v != "" {
		filters.LocationName = &v
	}

	page, pageSize := helpers.ParsePagination(c)

	requests, total, err := h.requestSvc.GetAllRequestsAdmin(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		log.Printf("ERROR GetAllRequestsAdmin: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.PaginatedResponse(c, http.StatusOK, requests, total, page, pageSize)
}
