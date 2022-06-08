package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	accountIDField = "accountid"
	profileField = "profile"
	identityStatusField = profileField + ".identitystatus"
)


type Mongo struct {
	col *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{col: db.Collection("profile")}
}

//数据的结构以以及字段
type ProfileRecord struct {
	AccountId string `bson:"accountid"`
	Profile *rentalpb.Profile `bson:"profile"`
}



//Get Profile
func (m *Mongo)GetProfile(c context.Context,aid id.AccountId)(*rentalpb.Profile,error)  {
	res := m.col.FindOne(c,byAccountID(aid))
	if err := res.Err(); err != nil {
		return nil, err
	}
	var pr ProfileRecord
	err := res.Decode(&pr)

	if err != nil {
		return nil,fmt.Errorf("cannot decode profile record:%v",err)
	}
	return pr.Profile,nil
}

func (m *Mongo)UpdateProfile(c context.Context,aid id.AccountId,s rentalpb.IdentityStatus,p *rentalpb.Profile) error  {
	_, err := m.col.UpdateOne(c, bson.M{
		accountIDField: aid.String(),
		identityStatusField:s,
	}, mgutil.Set(bson.M{
		profileField: p,
	}),options.Update().SetUpsert(true))//options.Update().SetUpsert(true)并不是普通的update，而是有则修改，无则新增
	return err
}

func byAccountID(aid id.AccountId) bson.M  {
	return bson.M{
		accountIDField: aid.String(),
	}
}






























