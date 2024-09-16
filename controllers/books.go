package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SigNoz/sample-golang-app/models"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type CreateBookInput struct {
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
}

type UpdateBookInput struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

// GET /books
// Find all books
func FindBooks(c *gin.Context) {
	var books []models.Book
	ctx := c.Request.Context()
	tracer := otel.Tracer("github.com/SigNoz/sample-golang-app/controllers")

	ctx, rootSpan := tracer.Start(ctx, "FindBooks-root")
	defer rootSpan.End()

	totalSpans := 20000
	batchSize := 1000
	for i := 0; i < totalSpans; i += batchSize {
		count := batchSize
		if i+batchSize > totalSpans {
			count = totalSpans - i
		}
		createFlatSpansIterative(ctx, tracer, count, i)

		// Force flush after each batch
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))
	}

	models.DB.WithContext(ctx).Find(&books)
	c.JSON(http.StatusOK, gin.H{"data": books})
}

func createFlatSpansIterative(ctx context.Context, tracer trace.Tracer, count, offset int) {
	for i := 0; i < count; i++ {
		_, span := tracer.Start(ctx, fmt.Sprintf("flat-span-%d", offset+i))
		span.SetAttributes(attribute.Int("numbered", offset+i))
		span.End()
		if (offset+i+1)%1000 == 0 {
			fmt.Printf("Created span %d\n", offset+i+1)
		}
	}
}

// GET /books/:id
// Find a book
func FindBook(c *gin.Context) {
	// Get model if exist
	var book models.Book
	if err := models.DB.WithContext(c.Request.Context()).Where("id = ?", c.Param("id")).First(&book).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": book})
}

// POST /books
// Create new book
func CreateBook(c *gin.Context) {
	// Validate input
	var input CreateBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create book
	book := models.Book{Title: input.Title, Author: input.Author}
	models.DB.WithContext(c.Request.Context()).Create(&book)

	c.JSON(http.StatusOK, gin.H{"data": book})
}

// PATCH /books/:id
// Update a book
func UpdateBook(c *gin.Context) {
	// Get model if exist
	var book models.Book
	if err := models.DB.WithContext(c.Request.Context()).Where("id = ?", c.Param("id")).First(&book).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// Validate input
	var input UpdateBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.WithContext(c.Request.Context()).Model(&book).Updates(input)

	c.JSON(http.StatusOK, gin.H{"data": book})
}

// DELETE /books/:id
// Delete a book
func DeleteBook(c *gin.Context) {
	// Get model if exist
	var book models.Book
	if err := models.DB.WithContext(c.Request.Context()).Where("id = ?", c.Param("id")).First(&book).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	models.DB.Delete(&book)

	c.JSON(http.StatusOK, gin.H{"data": true})
}
