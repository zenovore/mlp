package artifact

import (
	"context"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type gcsClient struct {
	API *storage.Client
}

type nopArtifact struct{}

type Config struct {
	Ctx context.Context
}

type Service interface {
	DeleteArtifact(url string, ctx context.Context) error
}

func NewGcsArtifactService(api *storage.Client) Service {
	return &gcsClient{
		API: api,
	}
}

func NewNopArtifactService() Service {
	return &nopArtifact{}
}

func (ad *nopArtifact) DeleteArtifact(url string, ctx context.Context) error {
	return nil
}
func (gc *gcsClient) DeleteArtifact(url string, ctx context.Context) error {
	// Get bucket name and gcsPrefix
	// the [5:] is to remove the "gs://" on the artifact uri
	// ex : gs://bucketName/path → bucketName/path
	gcsBucket, gcsLocation := gc.getGcsBucketAndLocation(url[5:])

	// Sets the name for the bucket.
	bucket := gc.API.Bucket(gcsBucket)

	it := bucket.Objects(ctx, &storage.Query{
		Prefix: gcsLocation,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (gc *gcsClient) getGcsBucketAndLocation(str string) (string, string) {
	// Split string using delimiter
	// ex : bucketName/path/path1/item → (bucketName , path/path1/item)
	splitStr := strings.SplitN(str, "/", 2)
	return splitStr[0], splitStr[1]
}
