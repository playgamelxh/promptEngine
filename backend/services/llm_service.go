package services

import (
	"bytes"
	"codeagent-backend/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LLMService struct{}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (s *LLMService) GenerateTestCases(ctx context.Context, config models.LLMConfig, promptContent string, count int) ([]models.TestCase, error) {
	// Construct the prompt to ask LLM to generate test cases (INPUTS ONLY)
	systemPrompt := `You are a QA engineer. Your task is to generate test inputs for a given prompt. 
Each test case must include:
1. "input": The test question or input data.

Return the result strictly as a JSON array of objects. 
Example format:
[
  {"input": "What is 2+2?"},
  {"input": "Translate 'Hello' to Spanish"}
]
Do not include any other text or markdown formatting.`
	userPrompt := fmt.Sprintf("Prompt: %s\n\nGenerate %d test inputs.", promptContent, count)

	response, err := s.CallLLM(ctx, config, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Clean up response (handle <think> tags, markdown, etc.)
	response = s.cleanAndExtractJSON(response)

	// Parse the response (expecting JSON array)
	var generatedInputs []struct {
		Input string `json:"input"`
	}
	err = json.Unmarshal([]byte(response), &generatedInputs)
	if err != nil {
		// Fallback: try to parse as strings if the model returned simple strings, or return error
		return nil, fmt.Errorf("failed to parse test cases: %v. Response: %s", err, response)
	}

	// Return test cases with inputs only (no expected output generated yet)
	var testCases []models.TestCase
	for _, gen := range generatedInputs {
		testCases = append(testCases, models.TestCase{
			Input: gen.Input,
		})
	}

	return testCases, nil
}

func (s *LLMService) GeneratePrompts(ctx context.Context, config models.LLMConfig, instruction string, count int) ([]models.Prompt, error) {
	systemPrompt := `You are a creative assistant. Your task is to generate prompts based on a user's instruction.
Each prompt must include:
1. "name": A short, descriptive name for the prompt.
2. "content": The actual prompt text.
3. "tags": Comma-separated tags (e.g., "coding,python,web").

Return the result strictly as a JSON array of objects.
Example format:
[
  {"name": "Python Helper", "content": "You are a Python expert...", "tags": "coding,python"},
  {"name": "Story Writer", "content": "Write a short story about...", "tags": "creative,writing"}
]
Do not include any other text or markdown formatting.`
	userPrompt := fmt.Sprintf("Instruction: %s\n\nGenerate %d prompts.", instruction, count)

	response, err := s.CallLLM(ctx, config, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	response = s.cleanAndExtractJSON(response)

	var generatedPrompts []struct {
		Name    string `json:"name"`
		Content string `json:"content"`
		Tags    string `json:"tags"`
	}
	err = json.Unmarshal([]byte(response), &generatedPrompts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prompts: %v. Response: %s", err, response)
	}

	var prompts []models.Prompt
	for _, gen := range generatedPrompts {
		prompts = append(prompts, models.Prompt{
			Name:    gen.Name,
			Content: gen.Content,
			Tags:    gen.Tags,
		})
	}

	return prompts, nil
}

func (s *LLMService) RunPrompt(ctx context.Context, config models.LLMConfig, promptContent string, input string) (string, error) {
	systemPrompt := "You are a helpful assistant."
	userPrompt := fmt.Sprintf("%s\n\nInput: %s", promptContent, input)
	return s.CallLLM(ctx, config, systemPrompt, userPrompt)
}

func (s *LLMService) EvaluateTestCase(ctx context.Context, config models.LLMConfig, promptContent string, input string, output string) (string, bool, error) {
	systemPrompt := "You are a QA engineer. Evaluate if the output matches the requirements of the prompt for the given input. Return a JSON object with 'is_pass' (boolean) and 'reason' (string). IMPORTANT: The 'reason' field MUST be written in the same language as the input text."
	userPrompt := fmt.Sprintf("Prompt: %s\nInput: %s\nOutput: %s", promptContent, input, output)

	response, err := s.CallLLM(ctx, config, systemPrompt, userPrompt)
	if err != nil {
		return "", false, err
	}

	response = s.cleanAndExtractJSON(response)

	var result struct {
		IsPass bool   `json:"is_pass"`
		Reason string `json:"reason"`
	}
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		return response, false, nil // Failed to parse, return raw response
	}

	return result.Reason, result.IsPass, nil
}

// cleanAndExtractJSON attempts to extract valid JSON from an LLM response
// It handles markdown code blocks, <think> tags (DeepSeek R1), and surrounding text.
func (s *LLMService) cleanAndExtractJSON(response string) string {
	// Remove <think>...</think> blocks (handling multiline)
	for {
		start := strings.Index(response, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(response, "</think>")
		if end == -1 {
			// Unclosed think tag? Just remove from start to end of string or keep it?
			// Better to just remove the tag itself if no closing, or maybe the model cut off.
			// For safety, if no closing tag, we might strip everything? Or just ignore.
			// Let's assume valid XML-like structure usually.
			break
		}
		// Remove the think block including tags
		response = response[:start] + response[end+8:]
	}

	response = strings.TrimSpace(response)

	// Remove markdown code blocks
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
	}
	response = strings.TrimSpace(response)

	// Attempt to find the first '{' or '[' and last '}' or ']'
	firstBrace := strings.IndexAny(response, "[{")
	lastBrace := strings.LastIndexAny(response, "]}")

	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		response = response[firstBrace : lastBrace+1]
	}

	return response
}

func (s *LLMService) CallLLM(ctx context.Context, config models.LLMConfig, systemContent string, userContent string) (string, error) {
	// Add 1 minute timeout for all LLM calls
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	// Handle Docker networking for localhost/127.0.0.1
	baseURL := config.BaseURL
	if strings.Contains(baseURL, "localhost") {
		baseURL = strings.Replace(baseURL, "localhost", "host.docker.internal", -1)
	}
	if strings.Contains(baseURL, "127.0.0.1") {
		baseURL = strings.Replace(baseURL, "127.0.0.1", "host.docker.internal", -1)
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Check if this is an Ollama native API request
	if strings.Contains(baseURL, "/api/generate") {
		return s.callOllamaNative(ctx, config, baseURL, systemContent, userContent)
	}

	reqBody := ChatRequest{
		Model: config.ModelName,
		Messages: []ChatMessage{
			{Role: "system", Content: systemContent},
			{Role: "user", Content: userContent},
		},
		Temperature: config.Temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// OpenAI compatible URL handling
	url := fmt.Sprintf("%s/chat/completions", baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (s *LLMService) callOllamaNative(ctx context.Context, config models.LLMConfig, url string, systemContent string, userContent string) (string, error) {
	// Combine system and user prompt for completion API
	fullPrompt := fmt.Sprintf("System: %s\n\nUser: %s", systemContent, userContent)

	reqBody := OllamaRequest{
		Model:  config.ModelName,
		Prompt: fullPrompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": config.Temperature,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	// Ollama usually doesn't need key, but we send it if present
	if config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+config.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}
