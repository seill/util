package util

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/oklog/ulid/v2"
	"gorm.io/datatypes"
)

func MapToStruct(m interface{}, s interface{}) (err error) {
	bytes, err := json.Marshal(m)
	if nil != err {
		return err
	}

	err = json.Unmarshal(bytes, s)

	return
}

func StructToMap(s interface{}) (m map[string]interface{}) {
	bytes, err := json.Marshal(s)
	if nil != err {
		m = map[string]interface{}{"error": err.Error()}
		return
	}

	err = json.Unmarshal(bytes, &m)
	if nil != err {
		m = map[string]interface{}{"error": err.Error()}
		return
	}

	return
}

func StructToString(s interface{}) string {
	bytes, err := json.Marshal(s)
	if nil != err {
		return ""
	}

	return string(bytes)
}

func MapToJson(m interface{}) (jsonData datatypes.JSON, err error) {
	bytes, err := json.Marshal(m)
	if nil != err {
		return
	}

	err = jsonData.Scan(bytes)

	return
}

func JsonToMap(jsonData datatypes.JSON) (m map[string]interface{}, err error) {
	bytes, err := jsonData.MarshalJSON()
	if nil != err {
		return
	}

	err = json.Unmarshal(bytes, &m)

	return
}

func GetJsonString(jsonData datatypes.JSON, key string, defaultValue string) (value string) {
	value = defaultValue

	dataBytes, _err := jsonData.MarshalJSON()
	if nil != _err {
		return
	}

	var dataMap map[string]interface{}
	_ = json.Unmarshal(dataBytes, &dataMap)

	if dataValue, ok := dataMap[key].(string); ok {
		value = dataValue
	}

	return
}

func GetJsonTimestamp(jsonData datatypes.JSON, key string, defaultValue time.Time) (value time.Time) {
	value = defaultValue

	dataBytes, _err := jsonData.MarshalJSON()
	if nil != _err {
		return
	}

	var dataMap map[string]interface{}
	_err = json.Unmarshal(dataBytes, &dataMap)
	if nil != _err {
		return
	}

	if dataValue, ok := dataMap[key].(string); ok {
		value, _ = time.Parse("2006-01-02T15:04:05Z", dataValue)
	}

	return
}

func GetJsonFloat64(jsonData datatypes.JSON, key string, defaultValue float64) (value float64) {
	value = defaultValue

	dataBytes, _err := jsonData.MarshalJSON()
	if nil != _err {
		return
	}

	var dataMap map[string]interface{}
	_err = json.Unmarshal(dataBytes, &dataMap)
	if nil != _err {
		return
	}

	if dataValue, ok := dataMap[key].(float64); ok {
		value = dataValue
	}

	return
}

func GetJsonInt64(jsonData datatypes.JSON, key string, defaultValue int64) (value int64) {
	value = defaultValue

	dataBytes, _err := jsonData.MarshalJSON()
	if nil != _err {
		return
	}

	var dataMap map[string]interface{}
	_err = json.Unmarshal(dataBytes, &dataMap)
	if nil != _err {
		return
	}

	if dataValue, ok := dataMap[key].(int64); ok {
		value = dataValue
	}

	return
}

func GetTimestampFormat(timestamp time.Time, layout string, location string) (date string) {
	date = "0000-00-00"

	loc, err := time.LoadLocation(location)
	if nil != err {
		return
	}

	date = timestamp.In(loc).Format(layout)

	return
}

type GetPresignedUrlRequest struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

type GetPresignedUrlResponse struct {
	Url string `json:"url"`
}

func GetPresignedUrlV2(request *GetPresignedUrlRequest) (response GetPresignedUrlResponse, err error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if nil != err {
		return
	}

	s3Client := s3.NewFromConfig(awsConfig)
	svc := s3.NewPresignClient(s3Client)

	request.Key, err = preProcessKey(request.Key)
	if nil != err {
		return
	}

	req, err := svc.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(request.Key),
	})

	if nil != err {
		return
	}

	response.Url = req.URL

	return
}

func preProcessKey(key string) (processedKey string, err error) {
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)

	splitKey := strings.Split(key, ".")
	if len(splitKey) > 1 {
		splitKey[len(splitKey)-2] = fmt.Sprintf("%s_%s", splitKey[len(splitKey)-2], id.String())
		processedKey = strings.Join(splitKey, ".")
	} else {
		processedKey = fmt.Sprintf("%s_%s", key, id.String())
	}

	return
}

func GetUlid() (id string) {
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	_id := ulid.MustNew(ulid.Timestamp(t), entropy)

	id = _id.String()

	return
}

func ConvertToAlphabetOnly(name string) (n string) {
	re, _ := regexp.Compile(`[^\w]`)

	n = re.ReplaceAllString(name, "")

	n = strings.ToUpper(strings.ReplaceAll(n, " ", ""))

	return
}

func GetDateStringByLocation(location string, format string, timestamp *time.Time) (date string) {
	loc, _ := time.LoadLocation(location)
	if nil == timestamp {
		timestamp = aws.Time(time.Now())
	}
	date = timestamp.In(loc).Format(format)

	return
}

func IsNil(v interface{}) bool {
	if v == nil {
		return true
	}

	switch val := reflect.ValueOf(v); val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return val.IsNil()
	default:
		return false
	}
}
