package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/internal/logger"
	"github.com/moulaybdl/incubAT/project_service/pkg/utils"
)

type MentorHandler struct {
	mentorService ports.MentorService
}

func NewMentorHandler(mentorService ports.MentorService) *MentorHandler {
	return &MentorHandler{
		mentorService: mentorService,
	}
}

type RegisterMentorRequest struct {
	Company          *string `json:"company"`
	Position         *string `json:"position"`
	ExpertiseArea    string  `json:"expertise_area"`
	YearsExperience  int     `json:"years_experience"`
	AvailabilityType string  `json:"availability_type"`
	LinkedinURL      *string `json:"linkedin_url"`
	WebsiteURL       *string `json:"website_url"`
	Phone            *string `json:"phone"`
}

func (h *MentorHandler) RegisterMentor(w http.ResponseWriter, r *http.Request) {
	var input RegisterMentorRequest

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Logger.Error("Failed to decode JSON in RegisterMentor", "error", err)
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid JSON format"}, nil)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		logger.Logger.Error("Failed to get user ID from context", "error", err)
		utils.WriteJSON(w, r, http.StatusUnauthorized, utils.Envelope{"error": "User not authenticated"}, nil)
		return
	}

	// Validate required fields
	if input.ExpertiseArea == "" {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "expertise_area is required"}, nil)
		return
	}

	if input.AvailabilityType == "" {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "availability_type is required"}, nil)
		return
	}

	if input.YearsExperience < 0 {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "years_experience cannot be negative"}, nil)
		return
	}

	// Check if user is already a mentor
	existingMentor, err := h.mentorService.GetMentorByUserID(userID)
	if err == nil && existingMentor != nil {
		logger.Logger.Warn("User is already registered as a mentor", "user_id", userID)
		utils.WriteJSON(w, r, http.StatusConflict, utils.Envelope{"error": "User is already registered as a mentor"}, nil)
		return
	}

	// Convert to service data structure
	mentorData := ports.MentorRegistrationData{
		Company:          input.Company,
		Position:         input.Position,
		ExpertiseArea:    input.ExpertiseArea,
		YearsExperience:  input.YearsExperience,
		AvailabilityType: input.AvailabilityType,
		LinkedinURL:      input.LinkedinURL,
		WebsiteURL:       input.WebsiteURL,
		Phone:            input.Phone,
	}

	// Register mentor via service (includes validation)
	mentor, err := h.mentorService.RegisterMentor(userID, mentorData)
	if err != nil {
		logger.Logger.Error("Failed to register mentor", "error", err, "user_id", userID)
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	logger.Logger.Info("Mentor registered successfully", "user_id", userID, "mentor_id", mentor.ID)

	utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{
		"success": true,
		"message": "Mentor registered successfully",
		"mentor":  mentor,
	}, nil)
}

// Helper function to get user ID from context
func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	return uuid.Parse(userID)
}
