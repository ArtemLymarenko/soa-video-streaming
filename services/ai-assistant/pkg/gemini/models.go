package gemini

import (
	"encoding/json"

	"google.golang.org/genai"
)

type ModelConfig struct {
	Name              string
	SystemInstruction string
	Temperature       *float32
	TopK              *int32
	TopP              *float32
	MaxOutputTokens   *int32
	Tools             []*Tool
}

type Tool struct {
	Name        string
	Description string
	Parameters  *ToolParameters
}

type ToolParameters struct {
	Type       string               `json:"type"`
	Properties map[string]*Property `json:"properties"`
	Required   []string             `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

func (t *Tool) ToGenaiTool() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        t.Name,
				Description: t.Description,
				Parameters: &genai.Schema{
					Type:       genai.TypeObject,
					Properties: t.toGenaiProperties(),
					Required:   t.Parameters.Required,
				},
			},
		},
	}
}

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

func (t *Tool) stringToType(typeStr string) genai.Type {
	switch typeStr {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeString
	}
}

type ToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

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

func (m *ModelConfig) WithSystemInstruction(instruction string) *ModelConfig {
	m.SystemInstruction = instruction
	return m
}

func (m *ModelConfig) WithTemperature(temp float32) *ModelConfig {
	m.Temperature = &temp
	return m
}

func (m *ModelConfig) WithTopK(topK int32) *ModelConfig {
	m.TopK = &topK
	return m
}

func (m *ModelConfig) WithTopP(topP float32) *ModelConfig {
	m.TopP = &topP
	return m
}

func (m *ModelConfig) WithMaxOutputTokens(tokens int32) *ModelConfig {
	m.MaxOutputTokens = &tokens
	return m
}

func (m *ModelConfig) AddTool(tool *Tool) *ModelConfig {
	m.Tools = append(m.Tools, tool)
	return m
}

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

func (t *Tool) AddParameter(name, paramType, description string) *Tool {
	t.Parameters.Properties[name] = &Property{
		Type:        paramType,
		Description: description,
	}
	return t
}

func (t *Tool) AddEnumParameter(name, description string, values ...string) *Tool {
	t.Parameters.Properties[name] = &Property{
		Type:        "string",
		Description: description,
		Enum:        values,
	}
	return t
}

func (t *Tool) MarkRequired(names ...string) *Tool {
	t.Parameters.Required = append(t.Parameters.Required, names...)
	return t
}
