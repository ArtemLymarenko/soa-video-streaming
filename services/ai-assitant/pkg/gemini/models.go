package gemini

import (
	"encoding/json"

	"github.com/google/generative-ai-go/genai"
)

// ModelConfig represents configuration for a Gemini model
type ModelConfig struct {
	Name              string
	SystemInstruction string
	Temperature       *float32
	TopK              *int32
	TopP              *float32
	MaxOutputTokens   *int32
	Tools             []*Tool
}

// Tool represents a function tool that can be called by the model
type Tool struct {
	Name        string
	Description string
	Parameters  *ToolParameters
}

// ToolParameters represents the parameters schema for a tool
type ToolParameters struct {
	Type       string               `json:"type"`
	Properties map[string]*Property `json:"properties"`
	Required   []string             `json:"required,omitempty"`
}

// Property represents a single parameter property
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// ToGenaiTool converts our Tool to genai.Tool
func (t *Tool) ToGenaiTool() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        t.Name,
				Description: t.Description,
				Parameters: &genai.Schema{
					Type:       "OBJECT",
					Properties: t.toGenaiProperties(),
					Required:   t.Parameters.Required,
				},
			},
		},
	}
}

// toGenaiProperties converts our properties to genai properties
func (t *Tool) toGenaiProperties() map[string]*genai.Schema {
	props := make(map[string]*genai.Schema)
	for key, prop := range t.Parameters.Properties {
		schema := &genai.Schema{
			Type:        t.stringToType(prop.Type),
			Description: prop.Description,
		}
		if len(prop.Enum) > 0 {
			schema.Enum = prop.Enum
		}
		props[key] = schema
	}
	return props
}

// stringToType converts string type to genai.Type
func (t *Tool) stringToType(typeStr string) string {
	switch typeStr {
	case "string":
		return "STRING"
	case "number":
		return "NUMBER"
	case "integer":
		return "INTEGER"
	case "boolean":
		return "BOOLEAN"
	case "array":
		return "ARRAY"
	case "object":
		return "OBJECT"
	default:
		return "STRING"
	}
}

// ToolCall represents a function call made by the model
type ToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// NewModelConfig creates a new model configuration with defaults
func NewModelConfig(name string) *ModelConfig {
	temp := float32(0.7)
	topK := int32(40)
	topP := float32(0.95)
	maxTokens := int32(2048)

	return &ModelConfig{
		Name:            name,
		Temperature:     &temp,
		TopK:            &topK,
		TopP:            &topP,
		MaxOutputTokens: &maxTokens,
		Tools:           make([]*Tool, 0),
	}
}

// WithSystemInstruction sets the system instruction for the model
func (m *ModelConfig) WithSystemInstruction(instruction string) *ModelConfig {
	m.SystemInstruction = instruction
	return m
}

// WithTemperature sets the temperature parameter
func (m *ModelConfig) WithTemperature(temp float32) *ModelConfig {
	m.Temperature = &temp
	return m
}

// WithTopK sets the TopK parameter
func (m *ModelConfig) WithTopK(topK int32) *ModelConfig {
	m.TopK = &topK
	return m
}

// WithTopP sets the TopP parameter
func (m *ModelConfig) WithTopP(topP float32) *ModelConfig {
	m.TopP = &topP
	return m
}

// WithMaxOutputTokens sets the max output tokens
func (m *ModelConfig) WithMaxOutputTokens(tokens int32) *ModelConfig {
	m.MaxOutputTokens = &tokens
	return m
}

// AddTool adds a tool to the model configuration
func (m *ModelConfig) AddTool(tool *Tool) *ModelConfig {
	m.Tools = append(m.Tools, tool)
	return m
}

// NewTool creates a new tool with the given name and description
func NewTool(name, description string) *Tool {
	return &Tool{
		Name:        name,
		Description: description,
		Parameters: &ToolParameters{
			Type:       "object",
			Properties: make(map[string]*Property),
			Required:   make([]string, 0),
		},
	}
}

// AddParameter adds a parameter to the tool
func (t *Tool) AddParameter(name, paramType, description string) *Tool {
	t.Parameters.Properties[name] = &Property{
		Type:        paramType,
		Description: description,
	}
	return t
}

// AddEnumParameter adds an enum parameter to the tool
func (t *Tool) AddEnumParameter(name, description string, values ...string) *Tool {
	t.Parameters.Properties[name] = &Property{
		Type:        "string",
		Description: description,
		Enum:        values,
	}
	return t
}

// MarkRequired marks a parameter as required
func (t *Tool) MarkRequired(names ...string) *Tool {
	t.Parameters.Required = append(t.Parameters.Required, names...)
	return t
}
