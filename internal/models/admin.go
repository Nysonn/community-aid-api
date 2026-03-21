package models

type DashboardStats struct {
	TotalUsers          int64   `json:"total_users"`
	ActiveUsers         int64   `json:"active_users"`
	TotalRequests       int64   `json:"total_requests"`
	PendingRequests     int64   `json:"pending_requests"`
	ApprovedRequests    int64   `json:"approved_requests"`
	RejectedRequests    int64   `json:"rejected_requests"`
	ClosedRequests      int64   `json:"closed_requests"`
	TotalOffers         int64   `json:"total_offers"`
	PendingOffers       int64   `json:"pending_offers"`
	AcceptedOffers      int64   `json:"accepted_offers"`
	FulfilledOffers     int64   `json:"fulfilled_offers"`
	TotalDonations      int64   `json:"total_donations"`
	TotalDonationAmount float64 `json:"total_donation_amount"`
}
