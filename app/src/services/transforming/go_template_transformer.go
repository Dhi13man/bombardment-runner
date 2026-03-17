package transforming

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	// text/template used intentionally: output is HTTP request payloads, not HTML.
	"text/template"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

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
) (GoTemplateTransformer, error) {
	transformer := goTemplateTransformer{clientChannel: clientChannel}

	if t, err := parseTemplate("body", transformerContext.BodyExpression); err != nil {
		return nil, err
	} else {
		transformer.bodyTemplate = t
	}
	if t, err := parseTemplate("endpoint", transformerContext.EndpointExpression); err != nil {
		return nil, err
	} else {
		transformer.endpointTemplate = t
	}
	if t, err := parseTemplate("headers", transformerContext.HeadersExpression); err != nil {
		return nil, err
	} else {
		transformer.headersTemplate = t
	}
	if t, err := parseTemplate("method", transformerContext.MethodExpression); err != nil {
		return nil, err
	} else {
		transformer.methodTemplate = t
	}

	return &transformer, nil
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
		// Parse as JSON so downstream json.Marshal round-trips correctly.
		// If the template output isn't valid JSON, use the raw string.
		var parsed interface{}
		if json.Unmarshal([]byte(result), &parsed) == nil {
			body = parsed
		} else {
			body = result
		}
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

func parseTemplate(name, text string) (*template.Template, error) {
	if text == "" {
		return nil, nil
	}
	// No custom functions registered. Built-in template functions (printf, len, etc.)
	// remain available but are safe since data values are plain strings.
	t, err := template.New(name).Funcs(template.FuncMap{}).Option("missingkey=error").Parse(text)
	if err != nil {
		return nil, fmt.Errorf("invalid Go template %q: %w", name, err)
	}
	return t, nil
}

func executeTemplate(t *template.Template, data map[string]string) (string, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// parseHeaderString parses a simple "Key: Value\nKey2: Value2" format into a map.
// Uses the first colon as separator, so header values may contain colons.
func parseHeaderString(raw string) map[string]string {
	headers := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r")
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			if key != "" {
				headers[key] = value
			}
		}
	}
	return headers
}
