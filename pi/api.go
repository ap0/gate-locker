package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIServer struct {
	controller *GateController
	config     *Config
}

type UnlockRequest struct {
	Duration string `json:"duration,omitempty"`
}

type StatusResponse struct {
	Message string `json:"message,omitempty"`
}

func NewAPIServer(controller *GateController, config *Config) *APIServer {
	return &APIServer{
		controller: controller,
		config:     config,
	}
}

func (api *APIServer) Start() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/status", api.handleStatus)
	router.POST("/unlock", api.handleUnlock)
	router.POST("/buttonPress", api.handleButtonPress)

	log.Printf("Starting API server on %s", api.config.APIConfig.Port)
	return router.Run(api.config.APIConfig.Port)
}

func (api *APIServer) handleStatus(c *gin.Context) {
	status := StatusResponse{
		Message: "Gate locker is running",
	}

	c.JSON(http.StatusOK, status)
}

func (a *APIServer) handleButtonPress(c *gin.Context) {
	if gpioMock, ok := a.controller.gpio.(*MockGPIOManager); ok {
		gpioMock.signalChan <- true
	}

	c.JSON(http.StatusOK, nil)
}

func (api *APIServer) handleUnlock(c *gin.Context) {
	api.controller.UnlockGate()

	response := StatusResponse{
		Message: "Unlock triggered",
	}

	c.JSON(http.StatusOK, response)
}
