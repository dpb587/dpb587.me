package cmdflags

import (
	"log/slog"

	"github.com/dpb587/tacitkb/catalog"
	storageos "github.com/dpb587/tacitkb/catalog/storage/os"
	"github.com/dpb587/tacitkb/ext/blob"
	"github.com/dpb587/tacitkb/tools/googlemapsreversegeocode"
)

type Global struct {
	Log        *slog.Logger
	Repository catalog.Repository

	FileService *storageos.Service
	BlobService *blob.Service

	ReverseGeo *googlemapsreversegeocode.Service
}
