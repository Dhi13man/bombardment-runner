package transforming

import models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"

type BaseTransformer interface {
	// Transforms the data based on the strategy.
	Transform(data map[string]string) (interface{}, error)

	// Returns the strategy to be used for the transformer.
	GetStrategy() models_enums.TransformerStrategy
}
