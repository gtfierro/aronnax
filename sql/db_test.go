package main

import (
	"flag"
	"fmt"
	"github.com/satori/go.uuid"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestMain(m *testing.M) {

	user := os.Getenv("ARONNAXTESTUSER")
	pass := os.Getenv("ARONNAXTESTPASS")
	dbname := os.Getenv("ARONNAXTESTDB")
	backend := newBackend(user, pass, dbname)
	uuid1, _ := uuid.FromString("2b365d6a-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid2, _ := uuid.FromString("370dd17c-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid3, _ := uuid.FromString("3a77a0e0-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid4, _ := uuid.FromString("3da1cafc-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid5, _ := uuid.FromString("411ce89c-8cbd-11e5-8bb3-0cc47a0f7eea")
	backend.RemoveData()

	initTags := map[string]string{
		"Location/City":            "Berkeley",
		"Location/Building":        "Soda",
		"Location/Floor":           "4",
		"Location/Room":            "410",
		"Properties/Timezone":      "America/Los_Angeles",
		"Properties/ReadingType":   "double",
		"Properties/UnitofMeasure": "F",
		"Properties/UnitofTime":    "ms",
		"Properties/StreamType":    "numeric",
	}

	temperatureTags := map[string]string{
		"Metadata/Point/Type":   "Sensor",
		"Metadata/Point/Sensor": "Temperature",
	}

	for i, doc := range []Document{ // initial documents
		Document{UUID: uuid1, Tags: initTags}, // 1
		Document{UUID: uuid2, Tags: initTags}, // 2
		Document{UUID: uuid3, Tags: initTags}, // 3
		Document{UUID: uuid4, Tags: initTags}, // 4
		Document{UUID: uuid5, Tags: initTags}, // 5

		// change location on uuid 1, 3, 5
		Document{UUID: uuid1, Tags: map[string]string{"Location/Room": "411"}}, // 6
		Document{UUID: uuid3, Tags: map[string]string{"Location/Room": "420"}}, // 7
		Document{UUID: uuid5, Tags: map[string]string{"Location/Room": "405"}}, // 8

		// add new tags describing temperature
		Document{UUID: uuid1, Tags: temperatureTags}, // 9
		Document{UUID: uuid2, Tags: temperatureTags}, // 10
		Document{UUID: uuid3, Tags: temperatureTags}, // 11
		Document{UUID: uuid4, Tags: temperatureTags}, // 12
		Document{UUID: uuid5, Tags: temperatureTags}, // 13

		// add exposure
		Document{UUID: uuid1, Tags: map[string]string{"Metadata/Exposure": "South"}}, // 14
		Document{UUID: uuid2, Tags: map[string]string{"Metadata/Exposure": "West"}},  // 15
		Document{UUID: uuid3, Tags: map[string]string{"Metadata/Exposure": "North"}}, // 16
		Document{UUID: uuid4, Tags: map[string]string{"Metadata/Exposure": "East"}},  // 17
		Document{UUID: uuid5, Tags: map[string]string{"Metadata/Exposure": "South"}}, // 18

		// delete exposure from one
		Document{UUID: uuid5, Tags: map[string]string{"Metadata/Exposure": ""}}, // 19
	} {
		// generate stricly ordered times so that we can write tests easily
		if err := backend.InsertWithTimestamp(&doc, time.Unix(int64(i)+1, 0)); err != nil {
			log.Fatal("Error inserting: %v", err)
		}
	}

	flag.Parse()
	os.Exit(m.Run())
}

func TestInsert(t *testing.T) {
	user := os.Getenv("ARONNAXTESTUSER")
	pass := os.Getenv("ARONNAXTESTPASS")
	dbname := os.Getenv("ARONNAXTESTDB")
	backend := newBackend(user, pass, dbname)
	uuid, _ := uuid.FromString("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae")

	for _, test := range []struct {
		doc Document
		ok  bool
	}{
		{
			Document{UUID: uuid, Tags: map[string]string{"key1": "val1", "key2": "val2"}},
			true,
		},
		{
			Document{UUID: uuid, Tags: map[string]string{"key1": ""}},
			true,
		},
	} {
		if err := backend.Insert(&test.doc); test.ok != (err == nil) {
			t.Errorf("Insert test failed: Expected err? %v Err: %v", test.ok, err)
		}
	}
}

// these tests run over the documents inserted in TestMain setup
func TestRecentDocument(t *testing.T) {
	user := os.Getenv("ARONNAXTESTUSER")
	pass := os.Getenv("ARONNAXTESTPASS")
	dbname := os.Getenv("ARONNAXTESTDB")
	backend := newBackend(user, pass, dbname)
	uuid1, _ := uuid.FromString("2b365d6a-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid2, _ := uuid.FromString("370dd17c-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid3, _ := uuid.FromString("3a77a0e0-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid4, _ := uuid.FromString("3da1cafc-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid5, _ := uuid.FromString("411ce89c-8cbd-11e5-8bb3-0cc47a0f7eea")

	for _, test := range []struct {
		uuid string
		doc  Document
	}{
		{
			uuid1.String(),
			Document{UUID: uuid1,
				Tags: map[string]string{
					"Location/City":            "Berkeley",
					"Location/Building":        "Soda",
					"Location/Floor":           "4",
					"Location/Room":            "411",
					"Properties/Timezone":      "America/Los_Angeles",
					"Properties/ReadingType":   "double",
					"Properties/UnitofMeasure": "F",
					"Properties/UnitofTime":    "ms",
					"Properties/StreamType":    "numeric",
					"Metadata/Point/Type":      "Sensor",
					"Metadata/Point/Sensor":    "Temperature",
					"Metadata/Exposure":        "South",
				},
			},
		},
		{
			uuid2.String(),
			Document{UUID: uuid2,
				Tags: map[string]string{
					"Location/City":            "Berkeley",
					"Location/Building":        "Soda",
					"Location/Floor":           "4",
					"Location/Room":            "410",
					"Properties/Timezone":      "America/Los_Angeles",
					"Properties/ReadingType":   "double",
					"Properties/UnitofMeasure": "F",
					"Properties/UnitofTime":    "ms",
					"Properties/StreamType":    "numeric",
					"Metadata/Point/Type":      "Sensor",
					"Metadata/Point/Sensor":    "Temperature",
					"Metadata/Exposure":        "West",
				},
			},
		},
		{
			uuid3.String(),
			Document{UUID: uuid3,
				Tags: map[string]string{
					"Location/City":            "Berkeley",
					"Location/Building":        "Soda",
					"Location/Floor":           "4",
					"Location/Room":            "420",
					"Properties/Timezone":      "America/Los_Angeles",
					"Properties/ReadingType":   "double",
					"Properties/UnitofMeasure": "F",
					"Properties/UnitofTime":    "ms",
					"Properties/StreamType":    "numeric",
					"Metadata/Point/Type":      "Sensor",
					"Metadata/Point/Sensor":    "Temperature",
					"Metadata/Exposure":        "North",
				},
			},
		},
		{
			uuid4.String(),
			Document{UUID: uuid4,
				Tags: map[string]string{
					"Location/City":            "Berkeley",
					"Location/Building":        "Soda",
					"Location/Floor":           "4",
					"Location/Room":            "410",
					"Properties/Timezone":      "America/Los_Angeles",
					"Properties/ReadingType":   "double",
					"Properties/UnitofMeasure": "F",
					"Properties/UnitofTime":    "ms",
					"Properties/StreamType":    "numeric",
					"Metadata/Point/Type":      "Sensor",
					"Metadata/Point/Sensor":    "Temperature",
					"Metadata/Exposure":        "East",
				},
			},
		},
		{
			uuid5.String(),
			Document{UUID: uuid5,
				Tags: map[string]string{
					"Location/City":            "Berkeley",
					"Location/Building":        "Soda",
					"Location/Floor":           "4",
					"Location/Room":            "405",
					"Properties/Timezone":      "America/Los_Angeles",
					"Properties/ReadingType":   "double",
					"Properties/UnitofMeasure": "F",
					"Properties/UnitofTime":    "ms",
					"Properties/StreamType":    "numeric",
					"Metadata/Point/Type":      "Sensor",
					"Metadata/Point/Sensor":    "Temperature",
				},
			},
		},
	} {
		query := fmt.Sprintf("select * where uuid = '%s';", test.uuid)
		if docs, err := DocsFromRows(backend.Eval(backend.Parse(query))); err != nil {
			t.Errorf("Query failed! %v", err)
		} else if len(docs) != 1 {
			t.Errorf("Only expected one doc! Got %v", len(docs))
		} else if !reflect.DeepEqual(test.doc, *(docs[0])) {
			t.Errorf("Does not match expected. Got\n%v\nwanted\n%v\n", docs[0], test.doc)
		}

	}
}

func TestWhereRecentDocument(t *testing.T) {
	user := os.Getenv("ARONNAXTESTUSER")
	pass := os.Getenv("ARONNAXTESTPASS")
	dbname := os.Getenv("ARONNAXTESTDB")
	backend := newBackend(user, pass, dbname)
	uuid1, _ := uuid.FromString("2b365d6a-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid2, _ := uuid.FromString("370dd17c-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid3, _ := uuid.FromString("3a77a0e0-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid4, _ := uuid.FromString("3da1cafc-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid5, _ := uuid.FromString("411ce89c-8cbd-11e5-8bb3-0cc47a0f7eea")

	for _, test := range []struct {
		querystring string             // query
		uuids       map[uuid.UUID]bool // expected matching UUIDs are keys. Initialize vals to false
	}{
		{
			"select distinct uuid where Location/Room = '410';",
			map[uuid.UUID]bool{uuid2: false, uuid4: false},
		},
		{
			"select distinct uuid where Location/Room != '410';",
			map[uuid.UUID]bool{uuid1: false, uuid3: false, uuid5: false},
		},
		{
			"select distinct uuid where Location/Room like '4%';",
			map[uuid.UUID]bool{uuid1: false, uuid2: false, uuid3: false, uuid4: false, uuid5: false},
		},
		{
			"select distinct uuid where has Location/Room;",
			map[uuid.UUID]bool{uuid1: false, uuid2: false, uuid3: false, uuid4: false, uuid5: false},
		},
		{
			"select distinct uuid where Metadata/Exposure = 'South' and has Properties/Timezone;",
			map[uuid.UUID]bool{uuid1: false},
		},
		{
			"select distinct uuid where (Metadata/Exposure = 'South' and has Properties/Timezone);",
			map[uuid.UUID]bool{uuid1: false},
		},
		{
			"select distinct uuid where Location/Room = '405' and has Metadata/Exposure;",
			map[uuid.UUID]bool{},
		},
		{
			"select distinct uuid where Location/Room = '405' or Location/Room = '411';",
			map[uuid.UUID]bool{uuid5: false, uuid1: false},
		},
		{
			"select distinct uuid where Location/Room = '405' or Location/Room = '411' or Location/Room = '420';",
			map[uuid.UUID]bool{uuid5: false, uuid3: false, uuid1: false},
		},
		{
			"select distinct uuid where Location/Room = '405' or (Location/Room = '411' and Metadata/Exposure='East');",
			map[uuid.UUID]bool{uuid5: false},
		},
		{
			"select distinct uuid where Location/Room = '405' or (Location/Room = '411' and Metadata/Exposure='South');",
			map[uuid.UUID]bool{uuid5: false, uuid1: false},
		},
		{
			"select distinct uuid where (Location/Room = '411' and Metadata/Exposure='South') or Location/Room = '405';",
			map[uuid.UUID]bool{uuid5: false, uuid1: false},
		},
		{
			"select distinct uuid where Location/Room = '405' or (Location/Room = '411' and (Location/City = 'Berkeley' and Metadata/Exposure='South'));",
			map[uuid.UUID]bool{uuid5: false, uuid1: false},
		},
		{
			"select distinct uuid where (Location/Room = '411' and (Location/City = 'Berkeley' and Metadata/Exposure='South')) or Location/Room = '405'; ",
			map[uuid.UUID]bool{uuid5: false, uuid1: false},
		},
	} {
		var (
			docs []*Document
			err  error
		)
		fmt.Println(test.querystring)
		if docs, err = DocsFromRows(backend.Eval(backend.Parse(test.querystring))); err != nil {
			t.Errorf("Query failed! %v", err)
		}
		for _, doc := range docs {
			if _, found := test.uuids[doc.UUID]; !found {
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
			} else {
				test.uuids[doc.UUID] = true
			}
		}

		for uuid, covered := range test.uuids {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}

func TestWhereWithNotRecentDocument(t *testing.T) {
	user := os.Getenv("ARONNAXTESTUSER")
	pass := os.Getenv("ARONNAXTESTPASS")
	dbname := os.Getenv("ARONNAXTESTDB")
	backend := newBackend(user, pass, dbname)
	//uuid1, _ := uuid.FromString("2b365d6a-8cbd-11e5-8bb3-0cc47a0f7eea")
	//uuid2, _ := uuid.FromString("370dd17c-8cbd-11e5-8bb3-0cc47a0f7eea")
	//uuid3, _ := uuid.FromString("3a77a0e0-8cbd-11e5-8bb3-0cc47a0f7eea")
	//uuid4, _ := uuid.FromString("3da1cafc-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid5, _ := uuid.FromString("411ce89c-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuiddummy, _ := uuid.FromString("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae")

	for _, test := range []struct {
		querystring string             // query
		uuids       map[uuid.UUID]bool // expected matching UUIDs are keys. Initialize vals to false
	}{
		{
			"select distinct uuid where not Location/Room like '4%';",
			map[uuid.UUID]bool{uuiddummy: false},
		},
		{
			"select distinct uuid where Location/Room = '405' and not has Metadata/Exposure;",
			map[uuid.UUID]bool{uuid5: false},
		},
	} {
		var (
			docs []*Document
			err  error
		)
		fmt.Println(test.querystring)
		if docs, err = DocsFromRows(backend.Eval(backend.Parse(test.querystring))); err != nil {
			t.Errorf("Query failed! %v", err)
		}
		for _, doc := range docs {
			if _, found := test.uuids[doc.UUID]; !found {
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
			} else {
				test.uuids[doc.UUID] = true
			}
		}

		for uuid, covered := range test.uuids {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}

func TestWhereWithTimePredicateWithBefore(t *testing.T) {
	user := os.Getenv("ARONNAXTESTUSER")
	pass := os.Getenv("ARONNAXTESTPASS")
	dbname := os.Getenv("ARONNAXTESTDB")
	backend := newBackend(user, pass, dbname)
	uuid1, _ := uuid.FromString("2b365d6a-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid2, _ := uuid.FromString("370dd17c-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid3, _ := uuid.FromString("3a77a0e0-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid4, _ := uuid.FromString("3da1cafc-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuid5, _ := uuid.FromString("411ce89c-8cbd-11e5-8bb3-0cc47a0f7eea")
	uuiddummy, _ := uuid.FromString("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae")
	for _, test := range []struct {
		querystring string             // query
		uuids       map[uuid.UUID]bool // expected matching UUIDs are keys. Initialize vals to false
	}{
		// BEFORE
		{
			"select distinct uuid where Location/Room = '410' before 5;",
			map[uuid.UUID]bool{uuid1: false, uuid2: false, uuid3: false, uuid4: false, uuid5: false},
		},
		{
			"select distinct uuid where Location/Room = '410' before 6;",
			map[uuid.UUID]bool{uuid1: false, uuid2: false, uuid3: false, uuid4: false, uuid5: false},
		},
		{
			"select distinct uuid where Location/Room = '410' before 8;",
			map[uuid.UUID]bool{uuid1: false, uuid2: false, uuid3: false, uuid4: false, uuid5: false},
		},
		{
			"select distinct uuid where Location/Room = '410' before 9;",
			map[uuid.UUID]bool{uuid1: false, uuid2: false, uuid3: false, uuid4: false, uuid5: false},
		},
		{
			"select distinct uuid where Location/Room = '411' before 5;",
			map[uuid.UUID]bool{},
		},
		{
			"select distinct uuid where Location/Room = '411' before 6;",
			map[uuid.UUID]bool{uuid1: false},
		},
		{
			"select distinct uuid where not has Metadata/Exposure before 13;",
			map[uuid.UUID]bool{uuid1: false, uuid2: false, uuid3: false, uuid4: false, uuid5: false, uuiddummy: false},
		},
		{
			"select distinct uuid where not has Metadata/Exposure before 14;",
			map[uuid.UUID]bool{uuid2: false, uuid3: false, uuid4: false, uuid5: false, uuiddummy: false},
		},
	} {
		var (
			docs []*Document
			err  error
		)
		fmt.Println(test.querystring)
		if docs, err = DocsFromRows(backend.Eval(backend.Parse(test.querystring))); err != nil {
			t.Errorf("Query failed! %v", err)
		}
		for _, doc := range docs {
			if _, found := test.uuids[doc.UUID]; !found {
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
			} else {
				test.uuids[doc.UUID] = true
			}
		}

		for uuid, covered := range test.uuids {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}
