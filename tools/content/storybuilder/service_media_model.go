package storybuilder

import (
	"context"
	"fmt"

	"github.com/dpb587/dpb587.me/tools/content"
	"github.com/dpb587/dpb587.me/tools/content/frontmatterparams"
	"github.com/dpb587/tacitkb/catalog"
	"github.com/dpb587/tacitkb/catalog/catalogutil"
	"github.com/dpb587/tacitkb/ext/blob"
	"github.com/dpb587/tacitkb/ext/export3dmodel"
	"github.com/dpb587/tacitkb/util/ptrutil"
)

func (b *Service) buildMediaModel(ctx context.Context, blobNode catalog.Node, blobProfile *blob.ProfileResourceData) (*content.Document, error) {
	templateData := &frontmatterparams.MediaType{
		CatalogNodeUID: ptrutil.Value(string(blobNode.UID())),
	}

	doc := &content.Document{
		Frontmatter: &content.Content_Frontmatter{
			Params: &content.Content_Frontmatter_Params{
				MediaType: templateData,
			},
			Type: &mediaString,
		},
	}

	{
		blobContentResource, err := b.repository.GetResource(ctx, blobNode.UID(), blob.ContentResourceDescriptor{})
		if err != nil {
			return nil, fmt.Errorf("get content: %w", err)
		}

		blobContentResourceContents, err := b.repository.GetResourceContentList(ctx, blobContentResource)
		if err != nil {
			return nil, fmt.Errorf("get content contents: %w", err)
		}

		if filepathName, ok := catalogutil.GetResourceContentFilepathBase(blobContentResourceContents); ok {
			doc.Frontmatter.Title = ptrutil.Value(filepathName)
		}
	}

	err := func() error {
		artifactsResource, err := b.repository.GetResource(ctx, blobNode.UID(), export3dmodel.ArtifactsResourceDescriptor{
			Profile: "default",
		}, catalog.RepositoryGetResourceConfig{
			Generate: true,
		})
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		artifactsData, err := export3dmodel.UnmarshalArtifactsResource(ctx, artifactsResource)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		templateData.ModelService = &frontmatterparams.MediaType_ModelService{
			ModelURL:   artifactsData.ModelURL,
			PosterURL:  artifactsData.PosterURL,
			ViewerHTML: artifactsData.ViewerHTML,
		}

		return nil
	}()
	if err != nil {
		return nil, fmt.Errorf("export3dmodel: %v", err)
	}

	return doc, nil
}
