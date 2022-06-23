package trip

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	poi2 "coolcar/rental/trip/client/poi"
	"coolcar/rental/trip/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/server"
	mongotesting "coolcar/shared/testing"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func TestTripLifecycle(t *testing.T) {
	c := auth.ContextWithAccountId(
		context.Background(),id.AccountId("account_for_lifecycle"))
	s := newService(c,t,&profileManager{},&carManager{})

	tid := id.TripId("dasfa69gs7g9h57sdf564f831")
	mgutil.NewObjIdWithValue(tid)
	cases := []struct{
		name string
		now int64
		op func()(*rentalpb.Trip,error)
		want string
	}{
		{
			name: "create_trip",
			now : 10000,
			op: func() (*rentalpb.Trip, error) {
				e,err := s.CreateTrip(c,&rentalpb.CreateTripRequest{
					Start: &rentalpb.Location{
						Latitude:  32.123,
						Longitude: 113.232,
					},
					CarId: "",
				})
				if err != nil {
					return nil, err
				}
				return e.Trip,nil
			},
		},
		{
			name: "update_trip",
			now: 20000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(c,&rentalpb.UpdateTripRequest{
					Id:      tid.String(),
					Current: &rentalpb.Location{
						Latitude: 28.21312,
						Longitude: 123.6342,
					},
					EndTrip: false,
				})
			},
		},
		{
			name: "finish_trip",
			now: 30000,
			op: func() (*rentalpb.Trip, error) {
				return s.UpdateTrip(c,&rentalpb.UpdateTripRequest{
					Id:      tid.String(),
					EndTrip: true,
				})
			},
		},
		{
			name: "query_trip",
			now:40000,
			op: func() (*rentalpb.Trip, error) {
				return s.GetTrip(c,&rentalpb.GetTripRequest{Id: tid.String()})
			},
		},
	}

	rand.Seed(1345)
	for _,cc := range cases {
		nowFunc = func() int64 {
			return cc.now
		}
		trip,err := cc.op()
		if err != nil {
			t.Errorf("%s:operation failed:%v",cc.name,err)
			continue
		}
		b,err := json.Marshal(trip)
		if err != nil {
			t.Errorf("%s:failed marshalling response:%v",cc.name,err)
		}
		got := string(b)
		t.Log(got)
	}


}

func newService(c context.Context,t *testing.T,pm ProfileManager,cm CarManager) *Service {
	//c := auth.ContextWithAccountId(context.Background(), id.AccountId("account1"))
	client, err := mongotesting.NewClient(c)
	if err != nil {
		t.Fatalf("cannot create mongo client: %v",err)
	}
	//pm := &profileManager{}
	//cm := &carManager{}

	logger,err := server.NewZapLogger()
	if err != nil {
		t.Fatalf("cannot create mongo client: %v",err)
	}

	db := client.Database("coolcar")
	err = mongotesting.SetupIndexes(c, db)

	return &Service{
		ProfileManager:                 pm,
		CarManager:                     cm,
		POIManager:                     &poi2.Manager{},
		Mongo: dao.NewMongo(db),
		Logger:logger,
		DistanceCalc: &distCalc{},
	}
}

type distCalc struct {
	
}

func (*distCalc) DistanceKm(c context.Context,from *rentalpb.Location,to *rentalpb.Location)(float64,error) {
	if from.Latitude == to.Latitude && from.Longitude == to.Longitude {
		return 0, nil
	}
	return 100, nil
}

//ACL层  防止入侵层(Anti Corruption Layer)
type profileManager struct {
	iID id.IdentityID
	err error
}
func (pm *profileManager)Verify(ctx context.Context,accountId id.AccountId) (id.IdentityID,error){

	return pm.iID, pm.err
}

//Car Manager defines the ACL for car management
type carManager struct {
	verfiErr error
	unlockErr error
}
func (cm *carManager)Verify(ctx context.Context,cid id.CarID,lc *rentalpb.Location) error{

	return cm.verfiErr
}
func (cm *carManager)Unlock(ctx context.Context,cid id.CarID,aid id.AccountId,tid id.TripId,avatarUrl string) error{

	return cm.unlockErr
}
func (cm *carManager)Lock(c context.Context,cid id.CarID) error  {
	return nil
}
//POIManager Point of Interest
type pOIManager struct {

}
func (pom *pOIManager)Resolve(ctx context.Context,lc *rentalpb.Location)(string,error){

	return "", nil
}

func TestService_CreateTrip(t *testing.T) {
	//c := auth.ContextWithAccountId(context.Background(), id.AccountId("account1"))
	c := context.Background()
	//client, err := mongotesting.NewClient(c)
	//if err != nil {
	//	t.Fatalf("cannot create mongo client: %v",err)
	//}
	pm := &profileManager{}
	cm := &carManager{}
	//
	//logger,err := server.NewZapLogger()
	//if err != nil {
	//	t.Fatalf("cannot create mongo client: %v",err)
	//}
	//
	//
	//s := &Service{
	//	ProfileManager:                 pm,
	//	CarManager:                     cm,
	//	POIManager:                     &poi2.Manager{},
	//	Mongo: dao.NewMongo(client.Database("coolcar")),
	//	Logger:logger,
	//}
	s := newService(c,t,pm,cm)

	nowFunc = func() int64 {
		return 1605695246
	}

	req := &rentalpb.CreateTripRequest{
		Start: &rentalpb.Location{
			Latitude:  32.123,
			Longitude: 113.232,
		},
		CarId: "",
	}

	pm.iID = "identity1"
	cases := []struct{
		name string
		accountId string
		tripId string
		profileErr error
		carVerifyErr error
		carUnlockErr error
		want string
		wantErr bool
	}{
		{
			name: "normal_create",
			tripId: "231231",
			want: "",
			accountId:"account1",
		},
		{
			name: "profile_err",
			tripId: "23123164g",
			profileErr: fmt.Errorf("profile"),
			wantErr: true,
			accountId: "account2",
		},
		{
			name: "car_verify_err",
			tripId: "23123164gvcx",
			profileErr: fmt.Errorf("verify"),
			wantErr: true,
			accountId: "account3",
		},
		{
			name: "car_unlock_err",
			tripId: "23123164gvcx123",
			profileErr: fmt.Errorf("unlock"),
			wantErr: false,
			accountId: "account4",
		},
	}
	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			mgutil.NewObjIdWithValue(id.TripId(cc.tripId))
			pm.err = cc.profileErr
			cm.unlockErr = cc.carUnlockErr
			cm.verfiErr = cc.carVerifyErr
			c = auth.ContextWithAccountId(context.Background(),id.AccountId(cc.accountId))
			trip, err := s.CreateTrip(c, req)
			if cc.wantErr {
				if err == nil {
					t.Errorf("want error : got none")
				}else{
					return
				}
			}
			if err != nil {
				t.Errorf("error creating trip: %v",err)
			}
			if trip.Id != cc.tripId {
				t.Errorf("incorrect id;want  %q,got %q",cc.tripId,trip.Id)
			}
			marshal, err := json.Marshal(trip.Trip)
			if err != nil {
				t.Errorf("cannot marshall response: %v",err)
			}
			tripStr := string(marshal)
			if cc.want != tripStr {
				t.Errorf("incorrect response: want %q,got %q",cc.want,tripStr)
			}
		})
	}
}

func TestMain(m *testing.M) {
	os.Exit(mongotesting.RunWithMongoInDocker(m))
}