package middlewares

import (
	"gopkg.in/mgo.v2"
	//"strings"
	//"time"
)

type backend struct {
	Db *mgo.Session
}

var Backend *backend

// func createDb(mgoInfo *mgo.DialInfo) (*mgo.Session, error) {
// 	mgoSess, err := mgo.DialWithInfo(mgoInfo)
// 	if err != nil {
// 		return nil, err
// 	}
// 	mgoSess.SetMode(mgo.Monotonic, true)
// 	return mgoSess, nil
// }

func InitBackend() error {
	mgoSess, err := mgo.Dial(Conf.Mongo.Addr)
	if err != nil {
		return err
	}
	mgoSess.SetMode(mgo.Monotonic, true)
	Backend = &backend{
		Db: mgoSess,
	}
	return nil
}
