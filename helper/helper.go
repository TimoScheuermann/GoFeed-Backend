package helper

import (
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func GetCurrentTimeMillies() int64 {
	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)
	return tUnixMilli
}

func CleanUpdateBody(body interface{}) *bson.M {
	return cleanBody(body, "remUpdate")
}
func CleanCreateBody(body interface{}) *bson.M {
	return cleanBody(body, "remInsert")
}

func cleanBody(body interface{}, tagVal string) *bson.M {
	val, typ := reflect.ValueOf(body), reflect.TypeOf(body)

	returnValue := bson.M{}

	for i := 0; i < val.NumField(); i++ {
		v, t := val.Field(i), typ.Field(i)

		tags := t.Tag.Get("gofeed")
		bso := t.Tag.Get("bson")
		if !strings.Contains(tags, tagVal) {
			returnValue[strings.Split(bso, ",")[0]] = v.Interface()
		}
	}
	return &returnValue
}
