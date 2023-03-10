package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/pkg/errors"
)

type OpenAI interface {
	QueryDeviceErrorCodes(make, model string, year int32, errorCodes []string) (string, error)
}

type openAi struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type ChatGptResponseChoices struct {
	Message struct {
		Role    string
		Content string
	}
	Index        int
	FinishReason string `json:"finish_reason"`
}

type ChatGptResponse struct {
	ID      string
	Object  string
	Created int
	Model   string
	Usage   struct {
		PromptTokens     int
		CompletionTokens int
		TotalTokens      int
	}
	Choices []ChatGptResponseChoices `json:"choices"`
}

func NewOpenAI(c config.Settings) OpenAI {
	return openAi{
		baseURL: c.OpenAiBaseURL,
		token:   c.OpenAiSecretKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (o openAi) askChatGpt(body *strings.Reader) (*ChatGptResponse, error) {
	var req *http.Request
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%schat/completions", o.baseURL),
		body,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.token))

	resp, err := http.DefaultClient.Do(req) // any error resp should return err per docs
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("received error from request: %s", string(b))
	}

	cResp := &ChatGptResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cResp)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding response json")
	}

	defer resp.Body.Close()

	return cResp, nil
}

func (o openAi) QueryDeviceErrorCodes(make, model string, year int32, errorCodes []string) (string, error) {
	codes := strings.Join(errorCodes[:], ",")
	req := fmt.Sprintf(`{
		"model": "gpt-3.5-turbo",
		"messages": [
			{
				"role": "user", 
				"content": "Briefly summarize for me in order what the following error codes mean for %s %s %d. The error codes are %s"}]
	  		}
	  `, make, model, year, codes)

	r, err := o.askChatGpt(strings.NewReader(req))
	if err != nil {
		return "", err
	}

	if len(r.Choices) < 1 {
		return "", nil
	}

	c := r.Choices[0]
	if c.FinishReason == "stop" {
		return strings.Trim(c.Message.Content, "\n"), nil
	}

	return "", nil
}
