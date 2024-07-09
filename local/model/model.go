package model

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type FilteredResponse struct {
	Items interface{} `json:"items"`
	Params
}

type ResponseFunc func(*fiber.Ctx) *[]interface{}

type MessageResponse struct {
	Message string `json:"message"`
}

type Params struct {
	SortBy       string `json:"sortBy"`
	Embed        string `json:"embed"`
	Page         int    `json:"page"`
	Rpp          int    `json:"rpp"`
	TotalRecords int    `json:"totalRecords"`
}

type BaseModel struct {
	Id          string    `json:"id"`
	DateCreated time.Time `json:"dateCreated"`
	DateUpdated time.Time `json:"dateUpdated"`
}

/*
Init

Initializes base model with DateCreated, DateUpdated, and Id values.
*/
func (cm *BaseModel) Init() {
	date := time.Now()
	cm.Id = uuid.NewString()
	cm.DateCreated = date
	cm.DateUpdated = date
}
