package bootstrap

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDriver struct {
	ClientCollection  *mongo.Collection
	Views             *mongo.Collection
	SessionSubscriber *mongo.Collection
	SessionDevice     *mongo.Collection
	Client            *mongo.Client
	Db                *mongo.Database
	Ctx               context.Context
}

func InitMongo(env *Env) DatabaseUseCase {
	db := MongoDriver{}
	db.Ctx = context.TODO()
	clientOptions := options.Client().ApplyURI(env.DB_HOST)
	client, err := mongo.Connect(db.Ctx, clientOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	db.Client = client
	db.Db = client.Database(env.DB_NAME)
	err = client.Ping(db.Ctx, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	db.ClientCollection = (*db.Db).Collection(env.CLIENT_COLLECTION)
	db.Views = (*db.Db).Collection(env.VIEWS_COLLECTION)
	db.SessionSubscriber = (*db.Db).Collection(env.SESSION_SUBSCRIBER_COLLECTION)
	db.SessionDevice = (*db.Db).Collection(env.SESSION_DEVICE_COLLECTION)
	return &db
}

func (instance *MongoDriver) Close() {
	instance.Db.Client().Disconnect(instance.Ctx)
}

// func (instance *MongoDriver) UpdateClientUUID(uuid string, arg ClientRecordUUID) (*ClientRecordUUID, error) {
// 	var upsert = true
// 	_, err := instance.ClientCollection.UpdateOne(
// 		instance.Ctx,
// 		bson.D{{Key: "uuid", Value: uuid}, {Key: "auth_id", Value: bson.M{"$exists": false}}},
// 		bson.M{"$set": arg},
// 		&options.UpdateOptions{Upsert: &upsert},
// 	)
// 	if err != nil {
// 		logrus.Printf("update client err: %s", err.Error())
// 	}
// 	client, err := instance.GetClientUUID(uuid)
// 	if err == nil {
// 		return client, nil
// 	}
// 	return nil, err
// }
