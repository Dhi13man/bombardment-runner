package transforming

import (
	"strings"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type PassthroughTransformer interface {
	BaseTransformer
}

type passthroughTransformer struct {
	clientChannel   modelsEnums.ClientChannel
	bodyColumns     []string          // column names to include, or ["*"] for all
	endpointMapping string            // column name or literal value
	methodMapping   string            // column name or literal value
	headerMappings  map[string]string // header-name -> column-name
}

func (pt *passthroughTransformer) GetStrategy() modelsEnums.TransformerStrategy {
	return modelsEnums.PASSTHROUGH
}

func NewPassthroughTransformer(
	clientChannel modelsEnums.ClientChannel,
	transformerContext modelsDtoTransforming.TransformerContext,
) PassthroughTransformer {
	transformer := passthroughTransformer{
		clientChannel: clientChannel,
	}

	if transformerContext.BodyExpression != "" {
		raw := strings.Split(transformerContext.BodyExpression, ",")
		cols := make([]string, 0, len(raw))
		for _, c := range raw {
			if trimmed := strings.TrimSpace(c); trimmed != "" {
				cols = append(cols, trimmed)
			}
		}
		transformer.bodyColumns = cols
	}

	transformer.endpointMapping = strings.TrimSpace(transformerContext.EndpointExpression)
	transformer.methodMapping = strings.TrimSpace(transformerContext.MethodExpression)

	// Format: "Header-Name=column_name,Another=col2"
	if transformerContext.HeadersExpression != "" {
		transformer.headerMappings = make(map[string]string)
		pairs := strings.Split(transformerContext.HeadersExpression, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) == 2 {
				transformer.headerMappings[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return &transformer
}

func (pt *passthroughTransformer) TransformRequest(data map[string]string) (
	modelsDtoRequests.BaseChannelRequest,
	error,
) {
	var body interface{}
	if len(pt.bodyColumns) > 0 {
		isWildcard := len(pt.bodyColumns) == 1 && pt.bodyColumns[0] == "*"
		size := len(pt.bodyColumns)
		if isWildcard {
			size = len(data)
		}
		bodyMap := make(map[string]interface{}, size)
		if isWildcard {
			for k, v := range data {
				bodyMap[k] = v
			}
		} else {
			for _, col := range pt.bodyColumns {
				if val, ok := data[col]; ok {
					bodyMap[col] = val
				} else {
					zap.L().Warn("Passthrough body column not found in data",
						zap.String("column", col),
					)
				}
			}
		}
		body = bodyMap
	}

	endpoint := pt.resolveMapping(pt.endpointMapping, data)
	method := pt.resolveMapping(pt.methodMapping, data)

	var headers map[string]string
	if len(pt.headerMappings) > 0 {
		headers = make(map[string]string, len(pt.headerMappings))
		for headerName, colName := range pt.headerMappings {
			headers[headerName] = pt.resolveMapping(colName, data)
		}
	}

	return createChannelRequest(pt.clientChannel, endpoint, body, headers, method)
}

func (pt *passthroughTransformer) resolveMapping(mapping string, data map[string]string) string {
	if mapping == "" {
		return ""
	}
	if val, ok := data[mapping]; ok {
		return val
	}
	return mapping
}
