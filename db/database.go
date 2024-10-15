package db

import (
	"context"
	"github.com/zgwit/iot-master/config"
	"github.com/zgwit/iot-master/pkg/exception"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var bucket *gridfs.Bucket

func Open() error {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().
		ApplyURI(config.GetString(MODULE, "url")).
		SetServerAPIOptions(serverAPI).
		SetBSONOptions(&options.BSONOptions{UseJSONStructTags: true})

	//鉴权
	auth := config.GetString(MODULE, "auth")
	username := config.GetString(MODULE, "username")
	password := config.GetString(MODULE, "password")
	if auth != "" && username != "" && password != "" {
		opts.SetAuth(options.Credential{
			AuthMechanism: "SCRAM-SHA-1",
			AuthSource:    auth,
			Username:      username,
			Password:      password,
		})
	}

	//连接
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return exception.Wrap(err)
	}

	db = client.Database(config.GetString(MODULE, "database"))

	//默认bucket files.fs files.chunks
	bucket, err = gridfs.NewBucket(db)
	if err != nil {
		_ = Close()
		return exception.Wrap(err)
	}

	return nil
}

func Ping() error {
	return db.Client().Ping(context.Background(), nil)
}

func Database() (*mongo.Database, error) {
	if db == nil {
		return nil, ErrDisconnect
	}
	return db, nil
}

func CreateTable(name string, opts *options.CreateCollectionOptions) error {
	if db == nil {
		return ErrDisconnect
	}
	err := db.CreateCollection(context.Background(), name, opts)
	if err != nil {
		return exception.Wrap(err)
	}
	return nil
}

func Tables() ([]string, error) {
	if db == nil {
		return nil, ErrDisconnect
	}
	ret, err := db.ListCollectionNames(context.Background(), bson.D{{}})
	return ret, exception.Wrap(err)
}

func Close() error {
	if db == nil {
		return ErrDisconnect
	}
	err := db.Client().Disconnect(context.TODO())
	db = nil
	bucket = nil
	return err
}
