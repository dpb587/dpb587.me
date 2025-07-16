package storybuilder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dpb587/dpb587.me/tools/content"
	"github.com/dpb587/tacitkb/catalog"
	"github.com/dpb587/tacitkb/ext/blob"
	"github.com/dpb587/tacitkb/tools/googlemapsreversegeocode"
)

type Service struct {
	log        *slog.Logger
	repository catalog.Repository
	rgeo       *googlemapsreversegeocode.Service
}

func NewService(log *slog.Logger, repository catalog.Repository, rgeo *googlemapsreversegeocode.Service) *Service {
	return &Service{
		log:        log,
		repository: repository,
		rgeo:       rgeo,
	}
}

func (b *Service) Build(ctx context.Context, blobNode catalog.Node) (*content.Document, error) {
	blobProfileResource, err := b.repository.GetResource(ctx, blobNode.UID(), blob.ProfileResourceDescriptor{}, catalog.RepositoryGetResourceConfig{
		Generate: true,
	})
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	blobProfile, err := blob.UnmarshalProfileResource(ctx, blobProfileResource)
	if err != nil {
		return nil, fmt.Errorf("unmarshal profile: %v", err)
	}

	//

	switch blobProfile.MediaType.ShortString() {
	case "application/gpx+xml":
		rb, err := b.buildRouteType(ctx, blobNode, blobProfile)
		if err != nil {
			return nil, fmt.Errorf("routeview: %w", err)
		}

		return rb, nil
	case "image/gif", "image/heif", "image/jpeg", "image/png":
		rb, err := b.buildMediaImage(ctx, blobNode, blobProfile)
		if err != nil {
			return nil, fmt.Errorf("imageview: %w", err)
		}

		return rb, nil
	}

	return nil, nil
}
