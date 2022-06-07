package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	//"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	tripField = "trip"
	accountIdField = tripField + ".accountid"
	statusField = tripField + ".status"
)

type Mongo struct {
	col *mongo.Collection
	//newObjId func() primitive.ObjectID
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		col:      db.Collection("trip"),
		//newObjId: primitive.NewObjectID,
	}
}

type TripRecord struct {
	mgutil.IDField       `bson:"inline"`
	mgutil.UpdateAtField `bson:"inline"`
	Trip                 *rentalpb.Trip `bson:"trip"`
}

//todo 同一个account 最多只能有一个进行中的Trip
//todo 强类型化tripID
//todo 表格驱动测试

func (m *Mongo)CreateTrip(c context.Context,trip *rentalpb.Trip)(*TripRecord,error)  {
	//var t TripRecord
	r := &TripRecord{
		Trip:          trip,
	}
	r.ID = mgutil.NewObjID()
	r.UpdateAt = mgutil.UpdateAt()
	_, err := m.col.InsertOne(c, r)
	if err != nil {
		return nil, err
	}
	return r,nil
}

func (m *Mongo)GetTrip(c context.Context,id id.TripId,accountId id.AccountId) (*TripRecord,error) {
	objId,err := objid.FromID(id)
	if err != nil {
		return nil, err
	}
	one := m.col.FindOne(c, bson.M{
		mgutil.IDFieldName: objId,
		accountIdField:     accountId,
	})
	if err := one.Err(); err != nil {
		return nil, err
	}

	var tr TripRecord
	err = one.Decode(&tr)
	if err != nil {
		return nil, err
	}
	return &tr,nil
}

func (m *Mongo)GetTrips(c context.Context,accountId id.AccountId,status rentalpb.TripStatus) ([]*TripRecord,error) {
	filter := bson.M{
		accountIdField: accountId.String(),
	}
	if status != rentalpb.TripStatus_TS_NOT_SPECIFIED {
		filter[statusField] = status
	}
	res ,err := m.col.Find(c,filter,options.Find().SetSort(bson.M{
		mgutil.IDFieldName: -1,//-1从大到小
	}))
	if err != nil {
		return nil, err
	}
	var trips []*TripRecord
	for res.Next(c) {
		var trip TripRecord
		err := res.Decode(&trip)
		if err != nil {
			return  nil,err
		}
		trips = append(trips,&trip)
	}
	return trips,nil
}


//Update Trip updates a trip ,乐观锁更新
func (m *Mongo)UpdateTrip(c context.Context,tid id.TripId,aid id.AccountId,updateAt int64,trip *rentalpb.Trip) error {
	objId,err := objid.FromID(tid)
	if err != nil {
		return err
	}

	newUpdateAt := mgutil.UpdateAt()

	res, err := m.col.UpdateOne(c, bson.M{
		mgutil.IDFieldName:       objId,
		accountIdField:           aid.String(),
		//mgutil.UpdateAtFieldName: updateAt,
	}, mgutil.Set(bson.M{
		tripField:                trip,
		mgutil.UpdateAtFieldName: newUpdateAt,
	}))

	if res == nil {
		return fmt.Errorf("update faild")
	}

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return  mongo.ErrNoDocuments
	}
	return nil
}