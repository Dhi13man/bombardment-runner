package clients

import (
	"errors"
	"io"

	"github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services"
)

type BaseChannelClient interface {
	services.BaseStrategy[modelsEnums.ClientChannel]
	io.Closer

	Execute(
		request modelsDtoRequests.BaseChannelRequest,
		baseUrl string,
	) (modelsDtoResponses.BaseChannelResponse, error)
}

func CreateChannelClient(
	context modelsDtoClients.ClientContext,
) (BaseChannelClient, error) {
	switch context.Channel {
	case modelsEnums.REST:
		return NewRestClient(context), nil
	case modelsEnums.GRAPHQL:
		return NewGraphqlClient(context), nil
	case modelsEnums.GRPC:
		return NewGrpcClient(context)
	default:
		return nil, errors.New("invalid strategy")
	}
}
