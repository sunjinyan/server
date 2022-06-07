package dao

import (
	"context"
	"coolcar/shared/id"
	mongo2 "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"

	//mgo "coolcar/shared"
	mongotesting "coolcar/shared/testing"
	"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
)
var mongoURI string
func TestMongo_ResolveAccountId(t *testing.T) {
	c := context.Background()
	//con, err := mongo.Connect(c, options.Client().ApplyURI("mongodb://47.93.20.75:27017/coolcar?readPreference=primary&ssl=false"))
	con, err := mongo.Connect(c, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("can not connect mongoDb: %v",err)
	}
	m := NewMongo(con.Database("coolcar"))
	_, err = m.col.InsertMany(c, []interface{}{
		bson.M{
			mongo2.IDFieldName: objid.MustFromId(id.AccountId("12312213")),
			openIDField:        "open_id_1",
		}, bson.M{
			mongo2.IDFieldName: objid.MustFromId(id.AccountId("12312213215")),
			openIDField:        "open_id_2",
		},
	})
	if err != nil {
		t.Fatalf("can not insert initial value:%v",err)
	}

	//m.newObjID = func() primitive.ObjectID {
	//	objID,_ := primitive.ObjectIDFromHex("12312213534")
	//	return objID
	//}
	/*mongo2.NewObjID = func() primitive.ObjectID {
		objID,_ := primitive.ObjectIDFromHex("12312213534")
		return objID
	}*/




	mongo2.NewObjIdWithValue(id.AccountId("12312213"))

	cases := []struct{
		name string
		openId string
		want string
	}{
		{
			name:"existing_user",
			openId: "openid_1",
			want:"12312213",
		},
		{
			name:"another_existing_user",
			openId: "openid_2",
			want:"12312213215",
		},
		{
			name:"new_user",
			openId: "openid_3",
			want:"12312213534",
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			ids,err := m.ResolveAccountId(context.Background(),
				cs.openId,
				)
			if err != nil {
				t.Errorf("can not insert initial value:%v",err)
			}
			t.Log("id == ",ids.String())
		})
	}
	ids, err := m.ResolveAccountId(c, "789")
	if err != nil {
		t.Fatalf("can not connect mongoDb: %v",err)
	}
	t.Log("id == ",ids.String())
}

//func mustObjID(hex string) primitive.ObjectID {
//	objID,err := primitive.ObjectIDFromHex(hex)
//	if err != nil {
//		panic(err)
//	}
//	return objID
//}

func TestMain(m *testing.M) {
	run := mongotesting.RunWithMongoInDocker(m)
	os.Exit(run)
}