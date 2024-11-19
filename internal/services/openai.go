package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type OpenAI interface {
	GetErrorCodesDescription(vMake, model string, errorCodes []string) ([]ErrorCodesResponse, error)
}

type openAI struct {
	chatGptURL string
	token      string
	httpClient *http.Client
	logger     *zerolog.Logger
}

type FunctionCallResponse struct {
	Name      string
	Arguments string
}

type ChatGPTResponseChoices struct {
	Message struct {
		Role         string
		Content      string               `json:"content"`
		FunctionCall FunctionCallResponse `json:"function_call"`
	}
	Index        int
	FinishReason string `json:"finish_reason"`
}

type ChatGPTResponse struct {
	ID      string
	Object  string
	Created int
	Model   string
	Usage   struct {
		PromptTokens     int
		CompletionTokens int
		TotalTokens      int
	}
	Choices []ChatGPTResponseChoices
}

type ErrorCodesResponse struct {
	Code        string `json:"code" example:"P0148"`
	Description string `json:"description" example:"Fuel delivery error"`
}

type ErrorCodesFunctionCallResponse struct {
	ErrorCodes []struct {
		Code        string
		Explanation string `json:"explanation"`
	} `json:"error_codes"`
}

func NewOpenAI(logger *zerolog.Logger, c config.Settings) OpenAI {
	return &openAI{
		chatGptURL: c.ChatGPTURL,
		token:      c.OpenAISecretKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (o *openAI) askChatGPT(body io.Reader) (*ChatGPTResponse, error) {
	ctx := context.Background()
	req, err := http.NewRequest(
		"POST",
		o.chatGptURL,
		body,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.token)

	start := time.Now()
	var currentReqResponseTime time.Duration
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			currentReqResponseTime = time.Since(start)
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	appmetrics.OpenAIResponseTimeOps.With(prometheus.Labels{
		"status": strconv.Itoa(resp.StatusCode),
	}).Observe(currentReqResponseTime.Seconds())
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("received error from request: %s", string(b))
	}

	cResp := &ChatGPTResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response json: %w", err)
	}

	return cResp, nil
}

func (o *openAI) GetErrorCodesDescription(vMake, model string, errorCodes []string) ([]ErrorCodesResponse, error) {
	codes := strings.Join(errorCodes, ", ")

	req := fmt.Sprintf(`{
		"model": "gpt-4o-mini",
		"temperature": 0,
		"messages": [
			{
				"role": "user", 
				"content": "A %s %s is returning error codes %s. Return a long extensive explanation for each code."
			}
		],
		"function_call": {
			"name": "vehicle_error_codes"
		},
		"functions": [
			{
				"name": "vehicle_error_codes",
				"parameters": {
					"type": "object",
					"properties": {
						"error_codes": {
							"type":"array",
							"items": {
								"type": "object",
								"properties": {
									"code": { "type": "string" },
									"explanation": { "type": "string" }
								},
								"required": ["code", "explanation"]
							}
						}
					},
					"required": ["error_codes"]
				}
			}
		]
	}
	`, vMake, model, codes)

	r, err := o.askChatGPT(strings.NewReader(req))
	if err != nil {
		return nil, errors.New("a temporary error occurred checking for your error codes, please try again")
	}

	appmetrics.OpenAITotalTokensUsedOps.Add(float64(r.Usage.TotalTokens))

	if len(r.Choices) == 0 {
		return nil, errors.New("could not fetch description for error codes")
	}

	c := r.Choices[0]
	if c.FinishReason != "stop" {
		o.logger.Error().Interface("rawResponse", r).Msg("Unexpected finish_reason from ChatGPT.")
	}

	var rawResp ErrorCodesFunctionCallResponse
	if err := json.Unmarshal([]byte(c.Message.FunctionCall.Arguments), &rawResp); err != nil {
		return nil, err
	}

	resp := []ErrorCodesResponse{}
	for _, obj := range rawResp.ErrorCodes {

		resp = append(resp, ErrorCodesResponse{
			Code:        obj.Code,
			Description: obj.Explanation,
		})
	}

	return resp, nil
}
