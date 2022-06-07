package dao

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	"coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/testing/protocmp"

	//mgo "coolcar/shared"
	mongotesting "coolcar/shared/testing"
	"os"
	"testing"
)

func TestMongo_CreateTrip(t *testing.T) {
	c := context.Background()
	con, err := mongotesting.NewDefaultClient(c)
	//con, err := mongo.Connect(c, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("can not connect mongoDb: %v",err)
	}
	db := con.Database("coolcar")
	err = mongotesting.SetupIndexes(c, db)
	if err != nil {
		t.Fatalf("cannot setup indexes: %v",err)
	}
	m := NewMongo(db)
	cases := []struct{
		name string
		tripID string
		accountID string
		tripStatus rentalpb.TripStatus
		wantErr bool
	}{
		{
			name : "finished",
			tripID:"123",
			accountID:"account1",
			tripStatus: rentalpb.TripStatus_FINISHED,
		},
		{
			name : "another_finished",
			tripID:"123222",
			accountID:"account1",
			tripStatus: rentalpb.TripStatus_FINISHED,
		},
		{
			name : "in_progress",
			tripID:"1232535622",
			accountID:"account1",
			tripStatus: rentalpb.TripStatus_IN_PROGRESS,
		},
		{
			name : "in_progress",
			tripID:"1232575635622",
			accountID:"account1",
			tripStatus: rentalpb.TripStatus_IN_PROGRESS,
			wantErr: true,
		},
		{
			name : "in_progress_by_another_account",
			tripID:"1232575sf635622",
			accountID:"account2",
			tripStatus: rentalpb.TripStatus_IN_PROGRESS,
		},
	}

	for _,cc := range cases{
		//mongo.NewObjID = func() primitive.ObjectID {
		//	return objid.MustFromId(id.TripId((cc.tripID)))
		//}
		mongo.NewObjIdWithValue(id.TripId(cc.tripID))
		tr,err := m.CreateTrip(c,&rentalpb.Trip{
			AccountId: cc.accountID,
			Status:    cc.tripStatus,
		})
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s : error expected got none",cc.name)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s :error create trip: %v",err,cc.name)
			continue
		}
		if tr.ID.Hex() != cc.tripID {
			t.Errorf("%s : incorrect trip id; want: %q got:%q",cc.name,cc.tripID ,tr.ID.Hex())
		}
	}
}


func TestMongo_GetTrip(t *testing.T) {
	//"mongodb://47.93.20.75:27017/coolcar?readPreference=primary&ssl=false"
	c := context.Background()
	con, err := mongotesting.NewDefaultClient(c)
	//con, err := mongo.Connect(c, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("can not connect mongoDb: %v",err)
	}
	m := NewMongo(con.Database("coolcar"))
	acct := id.AccountId("account1")
	mongo.NewObjID = primitive.NewObjectID
	trip, err := m.CreateTrip(c, &rentalpb.Trip{
		AccountId: acct.String(),
		CarId:     "car1",
		Start: &rentalpb.LocationStatus{
			Location: &rentalpb.Location{
				Latitude:  30,
				Longitude: 120,
			},
			PoiName: "startpoint",
		},
		End: &rentalpb.LocationStatus{
			Location: &rentalpb.Location{
				Latitude:  35,
				Longitude: 115,
			},
			PoiName:  "endpoint",
			FeeCent:  10000,
			KmDriven: 35,
		},
		Status: rentalpb.TripStatus_FINISHED,
	})
	if err != nil {
		t.Errorf("can not create trip : %v",err)
		return
	}
	t.Errorf("%+v",trip)

	getTrip, err := m.GetTrip(c,objid.ToTripId(trip.ID), acct)
	if err != nil {
		t.Errorf("can not get trip : %+v",err)
	}
	//t.Log(getTrip)
	if diff := cmp.Diff(trip,getTrip,protocmp.Transform()); diff != ""{
		t.Errorf("result differs : -want +got: %s",diff)
	}
}

func TestMongo_GetTrips(t *testing.T) {
	rows := []struct{
		id string
		accountId string
		status rentalpb.TripStatus
	}{
		{
			id:"35jlasd1wqde",
			accountId: "account_id_for_get_trips",
			status:rentalpb.TripStatus_FINISHED,
		},{
			id:"35jlasddasd",
			accountId: "account_id_for_get_trips",
			status:rentalpb.TripStatus_FINISHED,
		},{
			id:"35jlasdbxfbfx",
			accountId: "account_id_for_get_trips",
			status:rentalpb.TripStatus_FINISHED,
		},{
			id:"35jlasdrwtqq",
			accountId: "account_id_for_get_trips",
			status:rentalpb.TripStatus_FINISHED,
		},
	}

	//"mongodb://47.93.20.75:27017/coolcar?readPreference=primary&ssl=false"
	c := context.Background()
	con, err := mongotesting.NewDefaultClient(c)
	//con, err := mongo.Connect(c, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("can not connect mongoDb: %v",err)
	}
	m := NewMongo(con.Database("coolcar"))

	for _, r := range rows {
		mongo.NewObjIdWithValue(id.TripId(r.id))
		_, err := m.CreateTrip(c, &rentalpb.Trip{
			AccountId: r.accountId,
			Status:    r.status,
		})
		if err != nil {
			t.Fatalf("cannot create rows:%v",err)
		}
	}
	cases := []struct{
		name string
		accountId string
		status rentalpb.TripStatus
		wantCount int
		wantOnlyId  string
	}{
		{
			name: "get_all",
			accountId: "account_id_for_get_trips",
			status: rentalpb.TripStatus_TS_NOT_SPECIFIED,
			wantCount: 4,
		},{
			name: "get_in_progress",
			accountId: "account_id_for_get_trips",
			status: rentalpb.TripStatus_IN_PROGRESS,
			wantCount: 1,
			wantOnlyId: "35jlasdrwtqq",
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			trips, err := m.GetTrips(context.Background(), id.AccountId(cc.accountId), cc.status)
			if err != nil {
				t.Errorf("cannot get trips:%v",err)
			}
			if cc.wantCount != len(trips) {
				t.Errorf("want:%d len:%d",cc.wantCount,len(trips))
			}
			if cc.wantOnlyId != "" {
				if cc.wantOnlyId != trips[0].ID.Hex() {
					t.Errorf("want %q got %q",cc.wantOnlyId,trips[0].ID.Hex())
				}
			}
		})
	}
}

func TestMongo_UpdateTrip(t *testing.T) {
	c := context.Background()
	mc,err := mongotesting.NewClient(c)
	if err != nil {
		t.Fatalf("cannot connect mongodb: %v",err)
	}
	m := NewMongo(mc.Database("coolcar"))
	tid := id.TripId("1232575sf635622")
	aid := id.AccountId("account_for_update")

	var now int64 = 10000
	mongo.NewObjIdWithValue(tid)
	mongo.UpdateAt = func() int64 {
		return now
	}

	trip, err := m.CreateTrip(c, &rentalpb.Trip{
		AccountId: aid.String(),
		Status:    rentalpb.TripStatus_IN_PROGRESS,
		Start: &rentalpb.LocationStatus{
			PoiName:  "start_poi",
		},
	})
	if err != nil {
		t.Fatalf("cannot create trip: %v",err)
	}
	if trip.UpdateAt != 10000 {
		t.Fatalf("wrong updateat;want;10000,got:%d",trip.UpdateAt)
	}
	
	update := &rentalpb.Trip{
		AccountId: aid.String(),
		Status:    rentalpb.TripStatus_IN_PROGRESS,
		Start: &rentalpb.LocationStatus{
			PoiName:  "start_poi_update",
		},
	}
	
	cases := []struct{
		name string
		now int64
		withUpdateAt int64
		wantErr bool
	}{
		{
			name:"normal_update",
			now: 20000,
			withUpdateAt: 10000,
		},{
			name:"update_with_stale_timestamp",
			now: 30000,
			withUpdateAt: 10000,
			wantErr: true,
		},{
			name:"update_with_ref",
			now: 40000,
			withUpdateAt: 20000,
		},
	}

	for _, cc := range cases {
		now = cc.now
		err := m.UpdateTrip(c, tid, aid, cc.withUpdateAt, update)
		if cc.wantErr {
			if err == nil {
				t.Errorf("%s:want error;got none",cc.name)
			}else{
				continue
			}
		}else{
			if err != nil {
				t.Errorf("%s: cannot update:%v",cc.name,err)
			}
		}
		getTrip, err := m.GetTrip(c, tid, aid)
		if err != nil {
			t.Errorf("%s:cannot get trip after update:%v",cc.name,err)
		}
		if cc.now != getTrip.UpdateAt {
			t.Errorf("%s:incorrect updateat:   want %d,got %d",
				cc.name,cc.now,getTrip.UpdateAt)
		}
	}

}

func TestMain(m *testing.M) {
	run := mongotesting.RunWithMongoInDocker(m)
	os.Exit(run)
}