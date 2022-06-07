package dao

import (
	"context"
	"coolcar/shared/id"
	mongo2 "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"

	//mgo "coolcar/shared"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const openIDField = "open_id"

//Mongo defines a mongo dao.
type Mongo struct {
	col *mongo.Collection
	//newObjID func() primitive.ObjectID
}


//account collection
func NewMongo(db *mongo.Database)*Mongo  {
	return &Mongo{
		col: db.Collection("account"),
		//newObjID: primitive.NewObjectID,
	}
}

func (m *Mongo)ResolveAccountId(c context.Context,openId string)(id.AccountId,error)  {
	//m.col.InsertOne(c,bson.M{
	//	mgo.IDField: m.newObjID ,
	//	openIDField: openId,
	//})
	//insertedID := m.newObjID
	insertedID := mongo2.NewObjID()

	res :=  m.col.FindOneAndUpdate(c,bson.M{
		openIDField:openId,
	},  mongo2.SetOnInsert(bson.M{
		mongo2.IDFieldName: insertedID,
		openIDField:        openId,
	}),options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	if err := res.Err(); err != nil {
		return "", fmt.Errorf("can not findOneAndUpdate: %v",err)
	}

	var row mongo2.IDField

	err := res.Decode(&row)
	if err != nil {
		return "", fmt.Errorf("can not Decode: %v",err)
	}
	return objid.ToAccountId(row.ID),nil
}