package db

import (
	"context"
	"errors"
	"github.com/zgwit/iot-master/pkg/exception"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrDisconnect = errors.New("数据库未连接")

func Aggregate(tab string, pipeline any, results any) error {
	if db == nil {
		return ErrDisconnect
	}
	cursor, err := db.Collection(tab).Aggregate(context.Background(), pipeline)
	if err != nil {
		return exception.Wrap(err)
	}
	err = cursor.All(context.Background(), results)
	if err != nil {
		return exception.Wrap(err)
	}
	return nil
}

func BulkWrite(tab string, models []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	if db == nil {
		return nil, ErrDisconnect
	}
	ret, err := db.Collection(tab).BulkWrite(context.Background(), models)
	return ret, exception.Wrap(err)
}

func InsertOne(tab string, doc any) (id primitive.ObjectID, err error) {
	if db == nil {
		return primitive.NilObjectID, ErrDisconnect
	}
	ret, err := db.Collection(tab).InsertOne(context.Background(), doc)
	if err != nil {
		return primitive.NilObjectID, exception.Wrap(err)
	}
	return ParseObjectId(ret.InsertedID)
}

func InsertMany(tab string, docs []any) (ids []primitive.ObjectID, err error) {
	if db == nil {
		return nil, ErrDisconnect
	}
	ret, err := db.Collection(tab).InsertMany(context.Background(), docs)
	if err != nil {
		return nil, exception.Wrap(err)
	}
	for _, id := range ret.InsertedIDs {
		oid, err := ParseObjectId(id)
		if err == nil {
			ids = append(ids, oid)
		}
	}
	return ids, nil
}

func DeleteOne(tab string, filter any) (int64, error) {
	if db == nil {
		return 0, ErrDisconnect
	}
	ret, err := db.Collection(tab).DeleteOne(context.Background(), filter)
	if err != nil {
		return 0, exception.Wrap(err)
	}
	return ret.DeletedCount, nil
}

func DeleteMany(tab string, filter any) (int64, error) {
	if db == nil {
		return 0, ErrDisconnect
	}
	ret, err := db.Collection(tab).DeleteMany(context.Background(), filter)
	if err != nil {
		return 0, exception.Wrap(err)
	}
	return ret.DeletedCount, nil
}

func DeleteById(tab string, id primitive.ObjectID) (int64, error) {
	if db == nil {
		return 0, ErrDisconnect
	}
	ret, err := db.Collection(tab).DeleteOne(context.Background(), bson.D{{"_id", id}})
	if err != nil {
		return 0, exception.Wrap(err)
	}
	return ret.DeletedCount, nil
}

func ReplaceOne(tab string, filter any, result any, upsert bool) (int64, error) {
	if db == nil {
		return 0, ErrDisconnect
	}
	opts := options.Replace().SetUpsert(upsert)
	ret, err := db.Collection(tab).ReplaceOne(context.Background(), filter, result, opts)
	if err != nil {
		return 0, exception.Wrap(err)
	}
	return ret.ModifiedCount, nil
}

func UpdateOne(tab string, filter any, update any, upsert bool) (int64, error) {
	if db == nil {
		return 0, ErrDisconnect
	}
	opts := options.Update().SetUpsert(upsert)
	ret, err := db.Collection(tab).UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return 0, exception.Wrap(err)
	}
	return ret.ModifiedCount, nil
}

func UpdateMany(tab string, filter any, update any, upsert bool) (int64, error) {
	if db == nil {
		return 0, ErrDisconnect
	}
	opts := options.Update().SetUpsert(upsert)
	ret, err := db.Collection(tab).UpdateMany(context.Background(), filter, update, opts)
	if err != nil {
		return 0, exception.Wrap(err)
	}
	return ret.ModifiedCount, nil
}

func UpdateById(tab string, id primitive.ObjectID, update any, upsert bool) (int64, error) {
	if db == nil {
		return 0, ErrDisconnect
	}
	opts := options.Update().SetUpsert(upsert)
	ret, err := db.Collection(tab).UpdateByID(context.Background(), id, update, opts)
	if err != nil {
		return 0, exception.Wrap(err)
	}
	return ret.ModifiedCount, nil
}

func Find(tab string, filter any, sort any, skip int64, limit int64, results any) error {
	if db == nil {
		return ErrDisconnect
	}

	opts := options.Find()
	if sort != nil {
		opts.SetSort(sort)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}

	if filter == nil {
		filter = bson.D{{}}
	}

	ret, err := db.Collection(tab).Find(context.Background(), filter, opts)
	if err != nil {
		return exception.Wrap(err)
	}
	err = ret.All(context.Background(), results)
	return exception.Wrap(err)
}

func FindOne(tab string, filter any, result any) (has bool, err error) {
	if db == nil {
		return false, ErrDisconnect
	}
	err = db.Collection(tab).FindOne(context.Background(), filter).Decode(result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, exception.Wrap(err)
	}
	return true, nil
}

func FindOneAndDelete(tab string, filter any, raw any) (has bool, err error) {
	if db == nil {
		return false, ErrDisconnect
	}
	err = db.Collection(tab).FindOneAndDelete(context.Background(), filter).Decode(raw)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, exception.Wrap(err)
	}
	return true, nil
}

func FindOneAndUpdate(tab string, filter any, update any, raw any) (has bool, err error) {
	if db == nil {
		return false, ErrDisconnect
	}
	err = db.Collection(tab).FindOneAndUpdate(context.Background(), filter, update).Decode(raw)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, exception.Wrap(err)
	}
	return true, nil
}

func FindOneAndReplace(tab string, filter any, replace any, raw any) (has bool, err error) {
	if db == nil {
		return false, ErrDisconnect
	}
	err = db.Collection(tab).FindOneAndUpdate(context.Background(), filter, replace).Decode(raw)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, exception.Wrap(err)
	}
	return true, nil
}

func FindById(tab string, id primitive.ObjectID, result any) (has bool, err error) {
	if db == nil {
		return false, ErrDisconnect
	}
	err = db.Collection(tab).FindOne(context.Background(), bson.D{{"_id", id}}).Decode(result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, exception.Wrap(err)
	}
	return true, nil
}

func Count(tab string, filter any) (count int64, err error) {
	if db == nil {
		return 0, ErrDisconnect
	}
	ret, err := db.Collection(tab).CountDocuments(context.Background(), filter)
	return ret, exception.Wrap(err)
}

func Drop(tab string) error {
	if db == nil {
		return ErrDisconnect
	}
	err := db.Collection(tab).Drop(context.Background())
	return exception.Wrap(err)
}

func Distinct(tab string, filter any, field string) (values []any, err error) {
	if db == nil {
		return nil, ErrDisconnect
	}
	ret, err := db.Collection(tab).Distinct(context.Background(), field, filter)
	return ret, exception.Wrap(err)
}

func CreateIndex(tab string, keys []string) error {
	if db == nil {
		return ErrDisconnect
	}
	var ks bson.D
	for _, k := range keys {
		ks = append(ks, bson.E{Key: k, Value: 1}) //未支持降序索引
	}
	_, err := db.Collection(tab).Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    ks,
		Options: options.Index().SetSparse(true), //稀松索引
	})
	return exception.Wrap(err)
}

type _id struct {
	Id primitive.ObjectID `bson:"_id"`
}

func DistinctId(tab string, filter any) (ids []primitive.ObjectID, err error) {
	if db == nil {
		return nil, ErrDisconnect
	}

	opts := options.Find()
	if filter == nil {
		filter = bson.M{}
	}
	opts.Projection = bson.M{"_id": 1}

	ret, err := db.Collection(tab).Find(context.Background(), filter, opts)
	if err != nil {
		return nil, exception.Wrap(err)
	}
	var _ids []_id
	err = ret.All(context.Background(), &_ids)
	if err != nil {
		return nil, exception.Wrap(err)
	}
	for _, _id := range _ids {
		ids = append(ids, _id.Id)
	}
	return ids, nil
}
