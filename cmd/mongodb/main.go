package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	c := context.Background()
	mc, err := mongo.Connect(c, options.Client().ApplyURI("mongodb://47.93.20.75:27017/coolcar?readPreference=primary&ssl=false"))
	if err != nil {
		panic(err)
	}
	col := mc.Database("coolcar").Collection("account")
	res, err := col.InsertMany(c, []interface{}{
		bson.M{
			"open_id": "980",
		}, bson.M{
			"open_id": "760",
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v",res)
}
