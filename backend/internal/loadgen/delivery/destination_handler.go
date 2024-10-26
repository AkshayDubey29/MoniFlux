// backend/internal/loadgen/delivery/destination_handler.go

package delivery

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"bytes"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"github.com/sirupsen/logrus"
	"net/http"
)

// DestinationHandler defines the interface for handling different destinations.
type DestinationHandler interface {
	SendLog(log models.LogEntry) error
	SendMetric(metric models.Metric) error
	SendTrace(trace models.Trace) error
	Close() error
}

// HTTPDestinationHandler handles sending data to an HTTP endpoint.
type HTTPDestinationHandler struct {
	client   *http.Client
	endpoint string
	apiKey   string
	logger   *logrus.Logger
}

// NewHTTPDestinationHandler creates a new HTTPDestinationHandler.
func NewHTTPDestinationHandler(dest common.Destination, logger *logrus.Logger) *HTTPDestinationHandler {
	return &HTTPDestinationHandler{
		client:   &http.Client{Timeout: 10 * time.Second},
		endpoint: fmt.Sprintf("%s:%d", dest.Endpoint, dest.Port),
		apiKey:   dest.APIKey,
		logger:   logger,
	}
}

func (h *HTTPDestinationHandler) SendLog(log models.LogEntry) error {
	return h.sendPayload("logs", log)
}

func (h *HTTPDestinationHandler) SendMetric(metric models.Metric) error {
	return h.sendPayload("metrics", metric)
}

func (h *HTTPDestinationHandler) SendTrace(trace models.Trace) error {
	return h.sendPayload("traces", trace)
}

func (h *HTTPDestinationHandler) sendPayload(payloadType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		h.logger.Errorf("Failed to marshal %s payload: %v", payloadType, err)
		return err
	}

	url := fmt.Sprintf("http://%s/%s", h.endpoint, payloadType)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		h.logger.Errorf("Failed to create HTTP request for %s: %v", payloadType, err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if h.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.apiKey))
	}

	resp, err := h.client.Do(req)
	if err != nil {
		h.logger.Errorf("Failed to send %s to %s: %v", payloadType, h.endpoint, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.logger.Errorf("Received non-success status code %d when sending %s to %s", resp.StatusCode, payloadType, h.endpoint)
		return fmt.Errorf("non-success status code: %d", resp.StatusCode)
	}

	h.logger.Infof("Successfully sent %s to %s", payloadType, h.endpoint)
	return nil
}

func (h *HTTPDestinationHandler) Close() error {
	// No resources to clean up for HTTP
	return nil
}

// FileDestinationHandler handles writing data to local files.
type FileDestinationHandler struct {
	baseFilePath string
	fileCount    int
	fileFreq     int // Frequency in minutes
	currentFile  *os.File
	logger       *logrus.Logger
	ticker       *time.Ticker
	quit         chan struct{}
}

// NewFileDestinationHandler creates a new FileDestinationHandler.
func NewFileDestinationHandler(dest common.Destination, logger *logrus.Logger) (*FileDestinationHandler, error) {
	handler := &FileDestinationHandler{
		baseFilePath: dest.FilePath,
		fileCount:    dest.FileCount,
		fileFreq:     dest.FileFreq,
		logger:       logger,
		quit:         make(chan struct{}),
	}

	if err := handler.rotateFiles(); err != nil {
		return nil, err
	}

	handler.startFileRotation()
	return handler, nil
}

func (f *FileDestinationHandler) rotateFiles() error {
	if f.currentFile != nil {
		f.currentFile.Close()
	}

	timestamp := time.Now().Format("20060102_150405")
	newFilePath := fmt.Sprintf("%s_%s.log", f.baseFilePath, timestamp)
	file, err := os.Create(newFilePath)
	if err != nil {
		f.logger.Errorf("Failed to create log file %s: %v", newFilePath, err)
		return err
	}

	f.currentFile = file
	f.logger.Infof("Rotated to new log file: %s", newFilePath)
	return nil
}

func (f *FileDestinationHandler) startFileRotation() {
	duration := time.Duration(f.fileFreq) * time.Minute
	f.ticker = time.NewTicker(duration)

	go func() {
		for {
			select {
			case <-f.ticker.C:
				if err := f.rotateFiles(); err != nil {
					f.logger.Errorf("Error rotating files: %v", err)
				}
			case <-f.quit:
				f.ticker.Stop()
				if f.currentFile != nil {
					f.currentFile.Close()
				}
				return
			}
		}
	}()
}

func (f *FileDestinationHandler) SendLog(log models.LogEntry) error {
	return f.writeToFile(log)
}

func (f *FileDestinationHandler) SendMetric(metric models.Metric) error {
	return f.writeToFile(metric)
}

func (f *FileDestinationHandler) SendTrace(trace models.Trace) error {
	return f.writeToFile(trace)
}

func (f *FileDestinationHandler) writeToFile(payload interface{}) error {
	if f.currentFile == nil {
		return fmt.Errorf("log file not initialized")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		f.logger.Errorf("Failed to marshal payload: %v", err)
		return err
	}

	_, err = f.currentFile.Write(append(data, '\n'))
	if err != nil {
		f.logger.Errorf("Failed to write to file: %v", err)
		return err
	}

	f.logger.Debugf("Successfully wrote payload to file")
	return nil
}

func (f *FileDestinationHandler) Close() error {
	close(f.quit)
	return nil
}
