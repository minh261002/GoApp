package handler

import (
	"net/http"
	"strconv"
	"time"

	"go_app/internal/model"
	"go_app/internal/service"
	"go_app/pkg/response"

	"github.com/gin-gonic/gin"
)

// InventoryHandler handles inventory-related HTTP requests
type InventoryHandler struct {
	inventoryService service.InventoryService
}

// NewInventoryHandler creates a new InventoryHandler
func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{
		inventoryService: service.NewInventoryService(),
	}
}

// Inventory Movements

// CreateMovement creates a new inventory movement
func (h *InventoryHandler) CreateMovement(c *gin.Context) {
	var req model.InventoryMovementCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	movement, err := h.inventoryService.CreateMovement(&req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create inventory movement", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Inventory movement created successfully", movement)
}

// GetMovementByID retrieves an inventory movement by its ID
func (h *InventoryHandler) GetMovementByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid movement ID", err.Error())
		return
	}

	movement, err := h.inventoryService.GetMovementByID(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve movement", err.Error())
		return
	}

	if movement == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Movement not found", "movement not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Movement retrieved successfully", movement)
}

// GetMovements retrieves inventory movements with pagination and filters
func (h *InventoryHandler) GetMovements(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if productID := c.Query("product_id"); productID != "" {
		if id, err := strconv.ParseUint(productID, 10, 32); err == nil {
			filters["product_id"] = uint(id)
		}
	}
	if variantID := c.Query("variant_id"); variantID != "" {
		if id, err := strconv.ParseUint(variantID, 10, 32); err == nil {
			fariantID := uint(id)
			filters["variant_id"] = &fariantID
		}
	}
	if movementType := c.Query("type"); movementType != "" {
		filters["type"] = movementType
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if createdBy := c.Query("created_by"); createdBy != "" {
		if id, err := strconv.ParseUint(createdBy, 10, 32); err == nil {
			filters["created_by"] = uint(id)
		}
	}
	if reference := c.Query("reference"); reference != "" {
		filters["reference"] = reference
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if date, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters["date_from"] = date
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if date, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters["date_to"] = date
		}
	}

	movements, total, err := h.inventoryService.GetMovements(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve movements", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Movements retrieved successfully", movements, page, limit, total)
}

// UpdateMovement updates an existing inventory movement
func (h *InventoryHandler) UpdateMovement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid movement ID", err.Error())
		return
	}

	var req model.InventoryMovementUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	movement, err := h.inventoryService.UpdateMovement(uint(id), &req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update movement", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Movement updated successfully", movement)
}

// DeleteMovement deletes an inventory movement
func (h *InventoryHandler) DeleteMovement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid movement ID", err.Error())
		return
	}

	if err := h.inventoryService.DeleteMovement(uint(id)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete movement", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Movement deleted successfully", nil)
}

// ApproveMovement approves an inventory movement
func (h *InventoryHandler) ApproveMovement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid movement ID", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	movement, err := h.inventoryService.ApproveMovement(uint(id), userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to approve movement", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Movement approved successfully", movement)
}

// CompleteMovement completes an inventory movement
func (h *InventoryHandler) CompleteMovement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid movement ID", err.Error())
		return
	}

	movement, err := h.inventoryService.CompleteMovement(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to complete movement", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Movement completed successfully", movement)
}

// GetMovementsByProduct retrieves movements for a specific product/variant
func (h *InventoryHandler) GetMovementsByProduct(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var variantID *uint
	if variantIDStr := c.Query("variant_id"); variantIDStr != "" {
		if id, err := strconv.ParseUint(variantIDStr, 10, 32); err == nil {
			vID := uint(id)
			variantID = &vID
		}
	}

	movements, err := h.inventoryService.GetMovementsByProduct(uint(productID), variantID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve movements", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Movements retrieved successfully", movements)
}

// GetMovementsByReference retrieves movements by reference
func (h *InventoryHandler) GetMovementsByReference(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "Reference is required", "reference parameter is required")
		return
	}

	movements, err := h.inventoryService.GetMovementsByReference(reference)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve movements", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Movements retrieved successfully", movements)
}

// Stock Levels

// GetStockLevelByProduct retrieves stock level for a specific product/variant
func (h *InventoryHandler) GetStockLevelByProduct(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var variantID *uint
	if variantIDStr := c.Query("variant_id"); variantIDStr != "" {
		if id, err := strconv.ParseUint(variantIDStr, 10, 32); err == nil {
			vID := uint(id)
			variantID = &vID
		}
	}

	stockLevel, err := h.inventoryService.GetStockLevelByProduct(uint(productID), variantID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve stock level", err.Error())
		return
	}

	if stockLevel == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Stock level not found", "stock level not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Stock level retrieved successfully", stockLevel)
}

// GetAllStockLevels retrieves all stock levels with pagination and filters
func (h *InventoryHandler) GetAllStockLevels(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if productID := c.Query("product_id"); productID != "" {
		if id, err := strconv.ParseUint(productID, 10, 32); err == nil {
			filters["product_id"] = uint(id)
		}
	}
	if variantID := c.Query("variant_id"); variantID != "" {
		if id, err := strconv.ParseUint(variantID, 10, 32); err == nil {
			vID := uint(id)
			filters["variant_id"] = &vID
		}
	}
	if lowStock := c.Query("low_stock"); lowStock == "true" {
		filters["low_stock"] = true
	}
	if outOfStock := c.Query("out_of_stock"); outOfStock == "true" {
		filters["out_of_stock"] = true
	}

	stockLevels, total, err := h.inventoryService.GetAllStockLevels(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve stock levels", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Stock levels retrieved successfully", stockLevels, page, limit, total)
}

// UpdateStockLevelSettings updates stock level settings
func (h *InventoryHandler) UpdateStockLevelSettings(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var variantID *uint
	if variantIDStr := c.Query("variant_id"); variantIDStr != "" {
		if id, err := strconv.ParseUint(variantIDStr, 10, 32); err == nil {
			vID := uint(id)
			variantID = &vID
		}
	}

	var req model.StockLevelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	stockLevel, err := h.inventoryService.UpdateStockLevelSettings(uint(productID), variantID, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update stock level settings", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Stock level settings updated successfully", stockLevel)
}

// GetLowStockProducts retrieves products with low stock
func (h *InventoryHandler) GetLowStockProducts(c *gin.Context) {
	thresholdStr := c.DefaultQuery("threshold", "5")
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid threshold", err.Error())
		return
	}

	stockLevels, err := h.inventoryService.GetLowStockProducts(threshold)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve low stock products", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Low stock products retrieved successfully", stockLevels)
}

// GetOutOfStockProducts retrieves out of stock products
func (h *InventoryHandler) GetOutOfStockProducts(c *gin.Context) {
	stockLevels, err := h.inventoryService.GetOutOfStockProducts()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve out of stock products", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Out of stock products retrieved successfully", stockLevels)
}

// Inventory Adjustments

// CreateAdjustment creates a new inventory adjustment
func (h *InventoryHandler) CreateAdjustment(c *gin.Context) {
	var req model.InventoryAdjustmentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "user_id not found in context")
		return
	}

	adjustment, err := h.inventoryService.CreateAdjustment(&req, userID.(uint))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to create inventory adjustment", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "Inventory adjustment created successfully", adjustment)
}

// GetAdjustmentByID retrieves an inventory adjustment by its ID
func (h *InventoryHandler) GetAdjustmentByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid adjustment ID", err.Error())
		return
	}

	adjustment, err := h.inventoryService.GetAdjustmentByID(uint(id))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve adjustment", err.Error())
		return
	}

	if adjustment == nil {
		response.ErrorResponse(c, http.StatusNotFound, "Adjustment not found", "adjustment not found")
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Adjustment retrieved successfully", adjustment)
}

// GetAdjustments retrieves inventory adjustments with pagination and filters
func (h *InventoryHandler) GetAdjustments(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse filters
	filters := make(map[string]interface{})
	if productID := c.Query("product_id"); productID != "" {
		if id, err := strconv.ParseUint(productID, 10, 32); err == nil {
			filters["product_id"] = uint(id)
		}
	}
	if variantID := c.Query("variant_id"); variantID != "" {
		if id, err := strconv.ParseUint(variantID, 10, 32); err == nil {
			vID := uint(id)
			filters["variant_id"] = &vID
		}
	}
	if createdBy := c.Query("created_by"); createdBy != "" {
		if id, err := strconv.ParseUint(createdBy, 10, 32); err == nil {
			filters["created_by"] = uint(id)
		}
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if date, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters["date_from"] = date
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if date, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters["date_to"] = date
		}
	}

	adjustments, total, err := h.inventoryService.GetAdjustments(page, limit, filters)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve adjustments", err.Error())
		return
	}

	response.SuccessResponseWithPagination(c, http.StatusOK, "Adjustments retrieved successfully", adjustments, page, limit, total)
}

// GetAdjustmentsByProduct retrieves adjustments for a specific product/variant
func (h *InventoryHandler) GetAdjustmentsByProduct(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var variantID *uint
	if variantIDStr := c.Query("variant_id"); variantIDStr != "" {
		if id, err := strconv.ParseUint(variantIDStr, 10, 32); err == nil {
			vID := uint(id)
			variantID = &vID
		}
	}

	adjustments, err := h.inventoryService.GetAdjustmentsByProduct(uint(productID), variantID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve adjustments", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Adjustments retrieved successfully", adjustments)
}

// Statistics and Reports

// GetInventoryStats retrieves inventory statistics
func (h *InventoryHandler) GetInventoryStats(c *gin.Context) {
	stats, err := h.inventoryService.GetInventoryStats()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve inventory statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Inventory statistics retrieved successfully", stats)
}

// GetLowStockAlerts retrieves low stock alerts
func (h *InventoryHandler) GetLowStockAlerts(c *gin.Context) {
	alerts, err := h.inventoryService.GetLowStockAlerts()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve low stock alerts", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Low stock alerts retrieved successfully", alerts)
}

// GetStockValue calculates total stock value
func (h *InventoryHandler) GetStockValue(c *gin.Context) {
	value, err := h.inventoryService.GetStockValue()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to calculate stock value", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Stock value calculated successfully", gin.H{"total_value": value})
}

// GetMovementStats retrieves movement statistics for a date range
func (h *InventoryHandler) GetMovementStats(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", err.Error())
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default to 1 month ago
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", err.Error())
			return
		}
	} else {
		endDate = time.Now() // Default to now
	}

	stats, err := h.inventoryService.GetMovementStats(startDate, endDate)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve movement statistics", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Movement statistics retrieved successfully", stats)
}

// Stock Operations

// ReserveStock reserves stock for a product/variant
func (h *InventoryHandler) ReserveStock(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var variantID *uint
	if variantIDStr := c.Query("variant_id"); variantIDStr != "" {
		if id, err := strconv.ParseUint(variantIDStr, 10, 32); err == nil {
			vID := uint(id)
			variantID = &vID
		}
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.inventoryService.ReserveStock(uint(productID), variantID, req.Quantity); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to reserve stock", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Stock reserved successfully", nil)
}

// ReleaseStock releases reserved stock for a product/variant
func (h *InventoryHandler) ReleaseStock(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	var variantID *uint
	if variantIDStr := c.Query("variant_id"); variantIDStr != "" {
		if id, err := strconv.ParseUint(variantIDStr, 10, 32); err == nil {
			vID := uint(id)
			variantID = &vID
		}
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.inventoryService.ReleaseStock(uint(productID), variantID, req.Quantity); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to release stock", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Stock released successfully", nil)
}
