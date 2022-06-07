package objid

import (
	"coolcar/shared/id"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//func ToObjId(id fmt.Stringer) (primitive.ObjectID,error) {
func FromID(id fmt.Stringer) (primitive.ObjectID,error) {
	return primitive.ObjectIDFromHex(id.String())
}

func MustFromId(id fmt.Stringer) primitive.ObjectID {
	oid,err := FromID(id)
	if err != nil {
		panic(err)
	}
	return oid
}


func ToAccountId(oid primitive.ObjectID) id.AccountId  {
	return id.AccountId(oid.Hex())
}

func ToTripId(oid primitive.ObjectID) id.TripId  {
	return id.TripId(oid.Hex())
}