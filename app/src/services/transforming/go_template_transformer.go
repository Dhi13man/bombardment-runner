package transforming

import (
	"bytes"
	"fmt"
	"text/template"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type GoTemplateTransformer interface {
	BaseTransformer
}

type goTemplateTransformer struct {
	clientChannel    modelsEnums.ClientChannel
	bodyTemplate     *template.Template
	endpointTemplate *template.Template
	headersTemplate  *template.Template
	methodTemplate   *template.Template
}

func (gt *goTemplateTransformer) GetStrategy() modelsEnums.TransformerStrategy {
	return modelsEnums.GO_TEMPLATE
}

func NewGoTemplateTransformer(
	clientChannel modelsEnums.ClientChannel,
	transformerContext modelsDtoTransforming.TransformerContext,
) GoTemplateTransformer {
	transformer := goTemplateTransformer{clientChannel: clientChannel}

	if t := parseTemplateGracefully("body", transformerContext.BodyExpression); t != nil {
		transformer.bodyTemplate = t
	}
	if t := parseTemplateGracefully("endpoint", transformerContext.EndpointExpression); t != nil {
		transformer.endpointTemplate = t
	}
	if t := parseTemplateGracefully("headers", transformerContext.HeadersExpression); t != nil {
		transformer.headersTemplate = t
	}
	if t := parseTemplateGracefully("method", transformerContext.MethodExpression); t != nil {
		transformer.methodTemplate = t
	}

	return &transformer
}

func (gt *goTemplateTransformer) TransformRequest(data map[string]string) (
	modelsDtoRequests.BaseChannelRequest,
	error,
) {
	var body interface{}
	if gt.bodyTemplate != nil {
		result, err := executeTemplate(gt.bodyTemplate, data)
		if err != nil {
			return nil, fmt.Errorf("body template execution failed: %w", err)
		}
		body = result
	}

	var endpoint string
	if gt.endpointTemplate != nil {
		result, err := executeTemplate(gt.endpointTemplate, data)
		if err != nil {
			return nil, fmt.Errorf("endpoint template execution failed: %w", err)
		}
		endpoint = result
	}

	var headers map[string]string
	if gt.headersTemplate != nil {
		result, err := executeTemplate(gt.headersTemplate, data)
		if err != nil {
			return nil, fmt.Errorf("headers template execution failed: %w", err)
		}
		headers = parseHeaderString(result)
	}

	var method string
	if gt.methodTemplate != nil {
		result, err := executeTemplate(gt.methodTemplate, data)
		if err != nil {
			return nil, fmt.Errorf("method template execution failed: %w", err)
		}
		method = result
	}

	return createChannelRequest(gt.clientChannel, endpoint, body, headers, method)
}

func parseTemplateGracefully(name, text string) *template.Template {
	if text == "" {
		return nil
	}
	t, err := template.New(name).Option("missingkey=error").Parse(text)
	if err != nil {
		zap.L().Error("Error parsing Go template",
			zap.String("name", name),
			zap.Error(err),
		)
		return nil
	}
	return t
}

func executeTemplate(t *template.Template, data map[string]string) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// parseHeaderString parses a simple "Key: Value\nKey2: Value2" format into a map.
func parseHeaderString(raw string) map[string]string {
	headers := make(map[string]string)
	lines := splitLines(raw)
	for _, line := range lines {
		if idx := findColon(line); idx > 0 {
			key := trimSpace(line[:idx])
			value := trimSpace(line[idx+1:])
			if key != "" {
				headers[key] = value
			}
		}
	}
	return headers
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func findColon(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
