db.account.createIndex({
    "open_id":1,//从小到大
},{
    unique:true
})

db.trip.createIndex({
    "trip.accountid":1,//从小到大
    "trip.status":1//从小到大
},{
    unique:true,
    partialFilterExpression:{//
        "trip.status":1,//必须保证在status的值是1的情况下，唯一索引才会生效，0的不生效
    }
})

db.profile.createIndex({
    "accountid":1,
},{
    unique:true,
})