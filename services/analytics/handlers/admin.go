package handlers

import (
	"analytics-app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type AdminHandler struct {
	funnelRepo *repository.FunnelRepository
	eventRepo  *repository.EventRepository
	logger     zerolog.Logger
}

func NewAdminHandler(funnelRepo *repository.FunnelRepository, eventRepo *repository.EventRepository, logger zerolog.Logger) *AdminHandler {
	return &AdminHandler{
		funnelRepo: funnelRepo,
		eventRepo:  eventRepo,
		logger:     logger,
	}
}

// GetFunnelStats returns funnel statistics for admin dashboard
func (h *AdminHandler) GetFunnelStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get total funnel count
	totalFunnels, err := h.funnelRepo.GetTotalCount(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get total funnel count")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch funnel statistics"})
		return
	}

	// Get recent funnels
	recentFunnels, err := h.funnelRepo.GetRecentFunnels(ctx, 10)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get recent funnels")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent funnels"})
		return
	}

	// Get active funnel count
	activeFunnels, err := h.funnelRepo.GetActiveFunnelCount(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get active funnel count")
		activeFunnels = 0 // Default to 0 if error
	}

	stats := gin.H{
		"total":  totalFunnels,
		"recent": recentFunnels,
		"active": activeFunnels,
	}

	c.JSON(http.StatusOK, stats)
}

// GetAnalyticsStats returns analytics statistics for admin dashboard
func (h *AdminHandler) GetAnalyticsStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get total event count
	totalEvents, err := h.eventRepo.GetTotalEventCount(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get total event count")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics statistics"})
		return
	}

	// Get events today
	eventsToday, err := h.eventRepo.GetEventsToday(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get events today")
		eventsToday = 0 // Default to 0 if error
	}

	// Get unique visitors today
	uniqueVisitorsToday, err := h.eventRepo.GetUniqueVisitorsToday(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get unique visitors today")
		uniqueVisitorsToday = 0 // Default to 0 if error
	}

	// Get total pageviews
	totalPageviews, err := h.eventRepo.GetTotalPageviews(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get total pageviews")
		totalPageviews = 0 // Default to 0 if error
	}

	stats := gin.H{
		"total_events":          totalEvents,
		"events_today":          eventsToday,
		"unique_visitors_today": uniqueVisitorsToday,
		"total_pageviews":       totalPageviews,
	}

	c.JSON(http.StatusOK, stats)
}

// GetFunnelsList returns paginated list of funnels
func (h *AdminHandler) GetFunnelsList(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Get recent funnels (using existing method)
	funnels, err := h.funnelRepo.GetRecentFunnels(ctx, limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get funnels list")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch funnels list"})
		return
	}

	// Get total count
	totalFunnels, err := h.funnelRepo.GetTotalCount(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get total funnel count")
		totalFunnels = 0
	}

	totalPages := (totalFunnels + limit - 1) / limit // Ceiling division

	response := gin.H{
		"funnels": funnels,
		"pagination": gin.H{
			"current_page":   page,
			"total_pages":    totalPages,
			"total_funnels":  totalFunnels,
			"has_next_page":  page < totalPages,
			"has_prev_page":  page > 1,
		},
	}

	c.JSON(http.StatusOK, response)
}
