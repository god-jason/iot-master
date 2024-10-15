package db

import (
	"context"
	"github.com/zgwit/iot-master/pkg/exception"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
)

func Upload(filename string, metadata any) (*gridfs.UploadStream, error) {
	if bucket == nil {
		return nil, ErrDisconnect
	}
	opts := options.GridFSUpload().SetMetadata(metadata)
	return bucket.OpenUploadStream(filename, opts)
}

func UploadFrom(filename string, metadata any, reader io.Reader) (id primitive.ObjectID, err error) {
	if bucket == nil {
		return primitive.NilObjectID, ErrDisconnect
	}
	opts := options.GridFSUpload().SetMetadata(metadata)
	return bucket.UploadFromStream(filename, reader, opts)
}

func Download(filename string) (*gridfs.DownloadStream, error) {
	if bucket == nil {
		return nil, ErrDisconnect
	}
	return bucket.OpenDownloadStreamByName(filename)
}

func DownloadTo(filename string, writer io.Writer) (int64, error) {
	if bucket == nil {
		return 0, ErrDisconnect
	}
	return bucket.DownloadToStreamByName(filename, writer)
}

func DownloadById(id primitive.ObjectID) (*gridfs.DownloadStream, error) {
	if bucket == nil {
		return nil, ErrDisconnect
	}
	return bucket.OpenDownloadStream(id)
}

func DownloadToById(id primitive.ObjectID, writer io.Writer) (int64, error) {
	if bucket == nil {
		return 0, ErrDisconnect
	}
	return bucket.DownloadToStream(id, writer)
}

func Rename(id primitive.ObjectID, filename string) error {
	if bucket == nil {
		return ErrDisconnect
	}
	return bucket.Rename(id, filename)
}

func Remove(id primitive.ObjectID) error {
	if bucket == nil {
		return ErrDisconnect
	}
	return bucket.Delete(id)
}

func FindFile(filter any, sort any, skip int32, limit int32, results any) error {
	if bucket == nil {
		return ErrDisconnect
	}

	opts := options.GridFSFind()
	if sort != nil {
		opts.SetSort(sort)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}

	ret, err := bucket.Find(filter, opts)
	if err != nil {
		return exception.Wrap(err)
	}
	err = ret.All(context.Background(), results)
	return exception.Wrap(err)
}
