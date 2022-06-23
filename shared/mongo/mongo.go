package mongo

import (
	"coolcar/shared/mongo/objid"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const (
	IDFieldName = "_id"
	UpdateAtFieldName = "updatedat"
)

type IDField struct {
	ID primitive.ObjectID `bson:"_id"`
}

type UpdateAtField struct {
	UpdateAt int64 `bson:"updateat"`
}

var NewObjID = primitive.NewObjectID

func NewObjIdWithValue(id fmt.Stringer) {
	NewObjID = func() primitive.ObjectID {
		return objid.MustFromId(id)
	}
}


var UpdateAt = func() int64 {
	return time.Now().UnixNano()
}
//Set return a $set update document
func Set(v interface{}) bson.M {
	return bson.M{
		"$set":v,
	}
}
func SetOnInsert(v interface{}) bson.M {
	return bson.M{
		"$setOnInsert":v,
	}
}


//ZeroOrDoesNotExist
func ZeroOrDoesNotExist(filed string, zero interface{}) bson.M {
	return bson.M{
		"$or": []bson.M{
			{
				filed: zero,
			},
			{
				filed: bson.M{
					"$exists":false,
				},
			},
		},
	}
}