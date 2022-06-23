package dao

import (
	"context"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo struct {
	col *mongo.Collection
	//newObjId func() primitive.ObjectID
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		col:      db.Collection("blob"),
		//newObjId: primitive.NewObjectID,
	}
}

type BlobRecord struct {
	mgutil.IDField
	AccountID string `bson:"accountid"`
	Path string `bson:"path"`
}


//Create Blob Record
func (m *Mongo)CreateBlob(c context.Context,aid id.AccountId)(*BlobRecord,error)  {
	br := &BlobRecord{
		AccountID: aid.String(),
	}
	objId := mgutil.NewObjID()
	br.ID = objId

	br.Path = fmt.Sprintf("%s/%s",aid.String(),objId.Hex())

	_, err := m.col.InsertOne(c, br)

	if err != nil {
		return nil, err
	}

	return br, nil
}

func (m *Mongo)GetBlob(c context.Context,bid id.BlobID) (*BlobRecord,error) {
	objId,err := objid.FromID(bid)
	if err != nil {
		return nil,err
	}

	one := m.col.FindOne(c, bson.M{
		mgutil.IDFieldName: objId,
	})

	if err := one.Err(); err != nil {
		return nil, err
	}
	var br BlobRecord

	err = one.Decode(&br)
	if err != nil {
		return nil,err
	}

	return &br,nil
}



