package transforming

import (
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
	jsonata "github.com/blues/jsonata-go"
	"go.uber.org/zap"
)

type JsonataTransformer interface {
	BaseTransformer
}

type jsonataTransformer struct {
	expression string
}

func (jt *jsonataTransformer) GetStrategy() models_enums.TransformerStrategy {
	return models_enums.JSONATA
}

func NewJsonataTransformer(expression string) JsonataTransformer {
	return &jsonataTransformer{expression: expression}
}

func (jt *jsonataTransformer) Transform(data map[string]string) (interface{}, error) {
	if jt.expression == "" {
		return nil, nil
	}

	compiled, err := jsonata.Compile(jt.expression)
	if err != nil {
		zap.L().Error("Error compiling jsonata expression", zap.Error(err))
		return nil, err
	}
	return compiled.Eval(data)
}
