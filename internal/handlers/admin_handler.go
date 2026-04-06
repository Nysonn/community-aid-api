package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"community-aid-api/internal/helpers"
	"community-aid-api/internal/models"
	"community-aid-api/internal/services"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userSvc         *services.UserService
	requestSvc      *services.RequestService
	offerSvc        *services.OfferService
	donationSvc     *services.DonationService
	disbursementSvc *services.DisbursementService
	emailSvc        *services.EmailService
	db              *sql.DB
}

func NewAdminHandler(
	userSvc *services.UserService,
	requestSvc *services.RequestService,
	offerSvc *services.OfferService,
	donationSvc *services.DonationService,
	disbursementSvc *services.DisbursementService,
	emailSvc *services.EmailService,
	db *sql.DB,
) *AdminHandler {
	return &AdminHandler{
		userSvc:         userSvc,
		requestSvc:      requestSvc,
		offerSvc:        offerSvc,
		donationSvc:     donationSvc,
		disbursementSvc: disbursementSvc,
		emailSvc:        emailSvc,
		db:              db,
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

// GetDisbursements returns all disbursements, optionally filtered by status=pending|disbursed.
func (h *AdminHandler) GetDisbursements(c *gin.Context) {
	status := c.Query("status")
	page, pageSize := helpers.ParsePagination(c)

	disbursements, total, err := h.disbursementSvc.GetAllDisbursements(c.Request.Context(), status, page, pageSize)
	if err != nil {
		log.Printf("ERROR GetDisbursements: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.PaginatedResponse(c, http.StatusOK, disbursements, total, page, pageSize)
}

// MarkDisbursed marks a disbursement as sent and triggers emails to both donor and request owner.
func (h *AdminHandler) MarkDisbursed(c *gin.Context) {
	id := c.Param("id")

	d, err := h.disbursementSvc.MarkDisbursed(c.Request.Context(), id)
	if err == services.ErrNotFound {
		helpers.ErrorResponse(c, http.StatusNotFound, "disbursement not found")
		return
	}
	if err != nil {
		log.Printf("ERROR MarkDisbursed %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	// Fire-and-forget emails
	go h.sendDisbursementEmails(d)

	helpers.SuccessResponse(c, http.StatusOK, d)
}

func (h *AdminHandler) sendDisbursementEmails(d *models.Disbursement) {
	// Email to donor: funds delivered to recipient
	if err := h.emailSvc.SendDonorFundsDeliveredEmail(
		d.DonorEmail, d.RequestTitle, d.RecipientName, d.Amount,
	); err != nil {
		log.Printf("warn: failed to email donor %s: %v", d.DonorEmail, err)
	}

	// Email to request owner: funds sent to their account
	var ownerEmail string
	if err := h.db.QueryRowContext(context.Background(),
		`SELECT email FROM users WHERE id = (SELECT user_id FROM emergency_requests WHERE id = $1)`,
		d.RequestID,
	).Scan(&ownerEmail); err != nil {
		log.Printf("warn: could not fetch owner email for request %s: %v", d.RequestID, err)
		return
	}
	if err := h.emailSvc.SendFundsDisbursedToRecipientEmail(ownerEmail, d.RequestTitle, d.Amount); err != nil {
		log.Printf("warn: failed to email request owner %s: %v", ownerEmail, err)
	}
}

// PromoteToAdmin promotes a community_member to admin role.
func (h *AdminHandler) PromoteToAdmin(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userSvc.SetUserRole(c.Request.Context(), id, "admin")
	if err == services.ErrNotFound {
		helpers.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Printf("ERROR PromoteToAdmin %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, user)
}

// DemoteFromAdmin demotes an admin back to community_member.
func (h *AdminHandler) DemoteFromAdmin(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userSvc.SetUserRole(c.Request.Context(), id, "community_member")
	if err == services.ErrNotFound {
		helpers.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Printf("ERROR DemoteFromAdmin %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, user)
}
