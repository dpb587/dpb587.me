package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	var err error

	db, err = gorm.Open(sqlite.Open("main.sqlite"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// prototypeSample()

	jsonl := json.NewDecoder(os.Stdin)

	for {
		var recordRaw map[string]interface{}

		if err := jsonl.Decode(&recordRaw); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		record, err := buildRecord(recordRaw)
		if err != nil {
			panic(err)
		}

		table := reflect.ValueOf(record).Elem().FieldByName("Table").String()

		if err := db.Table(table).AutoMigrate(record); err != nil {
			panic(err)
		}

		if err := db.Table(table).Create(record).Error; err != nil {
			panic(err)
		}

		fmt.Printf("%#+v\n", record)
	}
}

var reValueRFC3339 = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)

func remarshal(out interface{}, in interface{}) error {
	buf, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshalling: %s", err)
	}

	err = json.Unmarshal(buf, &out)
	if err != nil {
		return fmt.Errorf("unmarshalling: %s", err)
	}

	return nil
}

func buildRecord(recordRaw map[string]interface{}) (interface{}, error) {
	recordType, err := buildType(recordRaw)
	if err != nil {
		return nil, fmt.Errorf("building type: %s", err)
	}

	record := reflect.New(recordType).Interface()

	if err := remarshal(&record, recordRaw); err != nil {
		return nil, fmt.Errorf("updating struct: %s", err)
	}

	return record, nil
}

func buildType(recordRaw map[string]interface{}) (reflect.Type, error) {
	var fields = []reflect.StructField{
		{
			Name: "Table",
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(`json:"@type" gorm:"-"`),
		},
	}

	for key, value := range recordRaw {
		var dbkey = key

		if value == nil {
			// cannot detect a field type without a value
			continue
		} else if key == "@type" {
			continue
		} else if key == "@id" {
			dbkey = "_id"
		}

		if valueT, ok := value.(string); ok && reValueRFC3339.MatchString(valueT) {
			valueTime, err := time.Parse(time.RFC3339, valueT)
			if err == nil {
				value = valueTime
			}
		}

		fields = append(fields, reflect.StructField{
			Name: fmt.Sprintf("Field%d", len(fields)),
			Type: reflect.TypeOf(value),
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s" gorm:"column:%s"`, key, dbkey)),
		})
	}

	return reflect.StructOf(fields), nil
}

func prototypeSample() {
	var sampleJSON = []byte(`{ "@type": "measurement",
  "@id": "a42fadde-ee76-4687-8f8f-303e083461e8",
  "at": "2020-10-29T23:17:21Z",
  "value": 79.3 }`)

	var dynamicFields = []reflect.StructField{
		{Name: "Table",
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(`json:"@type" gorm:"-"`)},
		{Name: "Field0",
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(`json:"@id" gorm:"column:_id;"`)},
		{Name: "Field1",
			Type: reflect.TypeOf(time.Time{}),
			Tag:  reflect.StructTag(`json:"at" gorm:"column:at"`)},
		{Name: "Field2",
			Type: reflect.TypeOf(float32(0)),
			Tag:  reflect.StructTag(`json:"value" gorm:"column:value"`)}}

	fmt.Printf("%#+v\n", sampleJSON)
	fmt.Printf("%#+v\n", dynamicFields)

	var dynamicType = reflect.StructOf(dynamicFields)
	var record = reflect.New(dynamicType).Interface()

	err := json.Unmarshal(sampleJSON, &record)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#+v\n", record)

	table := reflect.ValueOf(record).Elem().FieldByName("Table").String()
	fmt.Printf("%s\n", table)

	if err := db.Table(table).AutoMigrate(record); err != nil {
		panic(err)
	}

	if err := db.Table(table).Create(record).Error; err != nil {
		panic(err)
	}
}
