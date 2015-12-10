package main

import (
	"database/sql"
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
				ValidTime: MustParse(time.RFC3339, "1969-12-31T16:00:14Z"),
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
				ValidTime: MustParse(time.RFC3339, "1969-12-31T16:00:15Z"),
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
				ValidTime: MustParse(time.RFC3339, "1969-12-31T16:00:16Z"),
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
				ValidTime: MustParse(time.RFC3339, "1969-12-31T16:00:17Z"),
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
				ValidTime: MustParse(time.RFC3339, "1969-12-31T16:00:19Z"),
				//TODO: bug here. The test currently returns time 13, rather than 19. This is because it retrieves
				// the earliest version of the document equivalent to its latest form. Because it doesn't have an Exposure
				// tag at 19, it looks for the earliset time that it does, which is 13
			},
		},
	} {
		var (
			rows *sql.Rows
			docs []*Document
			err  error
		)
		query := fmt.Sprintf("select * where uuid = '%s';", test.uuid)
		if rows, _, err = backend.EvalWhere(backend.Parse(query)); err != nil {
			t.Errorf("Query failed! %v", err)
			continue
		}
		if docs, err = DocsFromRows(rows, ZERO_TIME); err != nil {
			t.Errorf("Doc transform failed! %v", err)
			continue
		}
		(docs[0].TagTimes) = nil
		if len(docs) != 1 {
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
		querystring string // query
		uuids       []uuid.UUID
	}{
		{
			"select distinct uuid where Location/Room = '410';",
			[]uuid.UUID{uuid2, uuid4},
		},
		{
			"select distinct uuid where Location/Room != '410';",
			[]uuid.UUID{uuid1, uuid3, uuid5},
		},
		{
			"select distinct uuid where Location/Room like '4%';",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where has Location/Room;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Metadata/Exposure = 'South' and has Properties/Timezone;",
			[]uuid.UUID{uuid1},
		},
		{
			"select distinct uuid where (Metadata/Exposure = 'South' and has Properties/Timezone);",
			[]uuid.UUID{uuid1},
		},
		{
			"select distinct uuid where Location/Room = '405' and has Metadata/Exposure;",
			[]uuid.UUID{},
		},
		{
			"select distinct uuid where Location/Room = '405' or Location/Room = '411';",
			[]uuid.UUID{uuid5, uuid1},
		},
		{
			"select distinct uuid where Location/Room = '405' or Location/Room = '411' or Location/Room = '420';",
			[]uuid.UUID{uuid5, uuid3, uuid1},
		},
		{
			"select distinct uuid where Location/Room = '405' or (Location/Room = '411' and Metadata/Exposure='East');",
			[]uuid.UUID{uuid5},
		},
		{
			"select distinct uuid where Location/Room = '405' or (Location/Room = '411' and Metadata/Exposure='South');",
			[]uuid.UUID{uuid5, uuid1},
		},
		{
			"select distinct uuid where (Location/Room = '411' and Metadata/Exposure='South') or Location/Room = '405';",
			[]uuid.UUID{uuid5, uuid1},
		},
		{
			"select distinct uuid where Location/Room = '405' or (Location/Room = '411' and (Location/City = 'Berkeley' and Metadata/Exposure='South'));",
			[]uuid.UUID{uuid5, uuid1},
		},
		{
			"select distinct uuid where (Location/Room = '411' and (Location/City = 'Berkeley' and Metadata/Exposure='South')) or Location/Room = '405'; ",
			[]uuid.UUID{uuid5, uuid1},
		},
	} {
		var (
			docs            []*Document
			rows            *sql.Rows
			expectedMatches = make(map[uuid.UUID]bool)
			err             error
		)
		for _, uid := range test.uuids {
			expectedMatches[uid] = false
		}
		if rows, _, err = backend.EvalWhere(backend.Parse(test.querystring)); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Query failed! %v", err)
			continue
		}
		if docs, err = DocsFromRows(rows, ZERO_TIME); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Doc transform failed! %v", err)
			continue
		}
		for _, doc := range docs {
			if _, found := expectedMatches[doc.UUID]; !found {
				fmt.Println(test.querystring)
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
				continue
			} else {
				expectedMatches[doc.UUID] = true
			}
		}

		for uuid, covered := range expectedMatches {
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
		querystring string // query
		uuids       []uuid.UUID
	}{
		{
			"select distinct uuid where not Location/Room like '4%';",
			[]uuid.UUID{uuiddummy},
		},
		{
			"select distinct uuid where Location/Room = '405' and not has Metadata/Exposure;",
			[]uuid.UUID{uuid5},
		},
	} {
		var (
			docs            []*Document
			rows            *sql.Rows
			expectedMatches = make(map[uuid.UUID]bool)
			err             error
		)
		for _, uid := range test.uuids {
			expectedMatches[uid] = false
		}
		if rows, _, err = backend.EvalWhere(backend.Parse(test.querystring)); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Query failed! %v", err)
			continue
		}
		if docs, err = DocsFromRows(rows, ZERO_TIME); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Doc transform failed! %v", err)
			continue
		}
		for _, doc := range docs {
			if _, found := expectedMatches[doc.UUID]; !found {
				fmt.Println(test.querystring)
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
				continue
			} else {
				expectedMatches[doc.UUID] = true
			}
		}

		for uuid, covered := range expectedMatches {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}

func TestWhereWithTimePredicateWithHappensBefore(t *testing.T) {
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
		querystring string // query
		uuids       []uuid.UUID
	}{
		// HAPPENS BEFORE
		{
			"select distinct uuid where Location/Room = '410' happens before 6;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room = '410' happens before 7;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room = '410' happens before 8;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room = '410' happens before 9;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room = '411' happens before 5;",
			[]uuid.UUID{},
		},
		{
			"select distinct uuid where Location/Room = '411' happens before 7;",
			[]uuid.UUID{uuid1},
		},
		{
			"select distinct uuid where not has Metadata/Exposure happens before 13;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5, uuiddummy},
		},
		{
			"select distinct uuid where not has Metadata/Exposure happens before 15;",
			[]uuid.UUID{uuid2, uuid3, uuid4, uuid5, uuiddummy},
		},
	} {
		var (
			docs            []*Document
			rows            *sql.Rows
			expectedMatches = make(map[uuid.UUID]bool)
			err             error
		)
		for _, uid := range test.uuids {
			expectedMatches[uid] = false
		}
		if rows, _, err = backend.EvalWhere(backend.Parse(test.querystring)); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Query failed! %v", err)
			continue
		}
		if docs, err = DocsFromRows(rows, ZERO_TIME); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Doc transform failed! %v", err)
			continue
		}
		for _, doc := range docs {
			if _, found := expectedMatches[doc.UUID]; !found {
				fmt.Println(test.querystring)
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
				continue
			} else {
				expectedMatches[doc.UUID] = true
			}
		}

		for uuid, covered := range expectedMatches {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}

func TestWhereWithTimePredicateWithBeforeWorkaround(t *testing.T) {
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
		querystring string // query
		uuids       []uuid.UUID
	}{
		// HAPPENS BEFORE OR AT
		{
			"select distinct uuid where Location/Room = '410' happens before 5 or Location/Room = '410' at 5;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room = '410' happens before 6 or Location/Room = '410' at 6;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room = '410' happens before 8 or Location/Room = '410' at 8;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room = '410' happens before 9 or Location/Room = '410' at 9;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room = '411' happens before 5 or Location/Room = '411' at 5;",
			[]uuid.UUID{},
		},
		{
			"select distinct uuid where Location/Room = '411' happens before 6 or Location/Room = '411' at 6;",
			[]uuid.UUID{uuid1},
		},
		{
			"select distinct uuid where not has Metadata/Exposure happens before 13 or has Metadata/Exposure at 13;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5, uuiddummy},
		},
		{
			"select distinct uuid where not has Metadata/Exposure happens before 14 or has Metadata/Exposure at 14;",
			[]uuid.UUID{uuid2, uuid3, uuid4, uuid5, uuiddummy},
		},
	} {
		var (
			docs            []*Document
			rows            *sql.Rows
			expectedMatches = make(map[uuid.UUID]bool)
			err             error
		)
		for _, uid := range test.uuids {
			expectedMatches[uid] = false
		}
		if rows, _, err = backend.EvalWhere(backend.Parse(test.querystring)); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Query failed! %v", err)
			continue
		}
		if docs, err = DocsFromRows(rows, ZERO_TIME); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Doc transform failed! %v", err)
			continue
		}
		for _, doc := range docs {
			if _, found := expectedMatches[doc.UUID]; !found {
				fmt.Println(test.querystring)
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
				continue
			} else {
				expectedMatches[doc.UUID] = true
			}
		}

		for uuid, covered := range expectedMatches {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}

func TestWhereWithTimePredicateWithAt(t *testing.T) {
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
		querystring string // query
		uuids       []uuid.UUID
	}{
		// AT
		{
			"select distinct uuid where Location/Room = '410' at 8;",
			[]uuid.UUID{uuid2, uuid4},
		},
		{
			"select distinct uuid where Location/Room = '410' at 6;",
			[]uuid.UUID{uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where not has Metadata/Exposure at 13;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5, uuiddummy},
		},
		{
			"select distinct uuid where not has Metadata/Exposure at 14;",
			[]uuid.UUID{uuid2, uuid3, uuid4, uuid5, uuiddummy},
		},
		{
			"select distinct uuid where not has Metadata/Exposure at 18;",
			[]uuid.UUID{uuiddummy},
		},
		{
			"select distinct uuid where not has Metadata/Exposure at 20;",
			[]uuid.UUID{uuiddummy, uuid5},
		},
	} {
		var (
			docs            []*Document
			rows            *sql.Rows
			expectedMatches = make(map[uuid.UUID]bool)
			err             error
		)
		for _, uid := range test.uuids {
			expectedMatches[uid] = false
		}
		if rows, _, err = backend.EvalWhere(backend.Parse(test.querystring)); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Query failed! %v", err)
			continue
		}
		if docs, err = DocsFromRows(rows, ZERO_TIME); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Doc transform failed! %v", err)
			continue
		}
		for _, doc := range docs {
			if _, found := expectedMatches[doc.UUID]; !found {
				fmt.Println(test.querystring)
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
				continue
			} else {
				expectedMatches[doc.UUID] = true
			}
		}

		for uuid, covered := range expectedMatches {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}

func TestWhereWithTimePredicateWithHappensAfter(t *testing.T) {
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
		querystring string // query
		uuids       []uuid.UUID
	}{
		// HAPPENS AFTER
		{
			"select distinct uuid where has Metadata/Exposure happens after 14;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where has Metadata/Exposure happens after 15;",
			[]uuid.UUID{uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where has Metadata/Exposure happens after 17;",
			[]uuid.UUID{uuid4, uuid5},
		},
		{
			"select distinct uuid where has Metadata/Exposure happens after 18;",
			[]uuid.UUID{uuid5},
		},
		{
			"select distinct uuid where has Metadata/Exposure happens after 19;",
			[]uuid.UUID{},
		},
	} {
		var (
			docs            []*Document
			rows            *sql.Rows
			expectedMatches = make(map[uuid.UUID]bool)
			err             error
		)
		for _, uid := range test.uuids {
			expectedMatches[uid] = false
		}
		if rows, _, err = backend.EvalWhere(backend.Parse(test.querystring)); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Query failed! %v", err)
			continue
		}
		if docs, err = DocsFromRows(rows, ZERO_TIME); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Doc transform failed! %v", err)
			continue
		}
		for _, doc := range docs {
			if _, found := expectedMatches[doc.UUID]; !found {
				fmt.Println(test.querystring)
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
				continue
			} else {
				expectedMatches[doc.UUID] = true
			}
		}

		for uuid, covered := range expectedMatches {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}

func TestWhereWithTimePredicateWithHappensIn(t *testing.T) {
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
		querystring string // query
		uuids       []uuid.UUID
	}{
		// HAPPENS IN
		{
			"select distinct uuid where Location/Room='410' happens in (1, 6);",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room='410' happens in (1, 8);",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room='410' happens in (6, 8);",
			[]uuid.UUID{},
		},
	} {
		var (
			docs            []*Document
			rows            *sql.Rows
			expectedMatches = make(map[uuid.UUID]bool)
			err             error
		)
		for _, uid := range test.uuids {
			expectedMatches[uid] = false
		}
		if rows, _, err = backend.EvalWhere(backend.Parse(test.querystring)); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Query failed! %v", err)
			continue
		}
		if docs, err = DocsFromRows(rows, ZERO_TIME); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Doc transform failed! %v", err)
			continue
		}
		for _, doc := range docs {
			if _, found := expectedMatches[doc.UUID]; !found {
				fmt.Println(test.querystring)
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
				continue
			} else {
				expectedMatches[doc.UUID] = true
			}
		}

		for uuid, covered := range expectedMatches {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}

func TestWhereWithTimePredicateWithHappensInWorkaround(t *testing.T) {
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
		querystring string // query
		uuids       []uuid.UUID
	}{
		// HAPPENS IN OR AT
		{
			"select distinct uuid where Location/Room='410' happens in (1, 6) or Location/Room='410' at 1;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room='410' happens in (1, 8) or Location/Room='410' at 1;",
			[]uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5},
		},
		{
			"select distinct uuid where Location/Room='410' happens in (6, 8) or Location/Room='410' at 6;",
			[]uuid.UUID{uuid2, uuid3, uuid4, uuid5},
		},
	} {
		var (
			docs            []*Document
			rows            *sql.Rows
			expectedMatches = make(map[uuid.UUID]bool)
			err             error
		)
		for _, uid := range test.uuids {
			expectedMatches[uid] = false
		}
		if rows, _, err = backend.EvalWhere(backend.Parse(test.querystring)); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Query failed! %v", err)
			continue
		}
		if docs, err = DocsFromRows(rows, ZERO_TIME); err != nil {
			fmt.Println(test.querystring)
			t.Errorf("Doc transform failed! %v", err)
			continue
		}
		for _, doc := range docs {
			if _, found := expectedMatches[doc.UUID]; !found {
				fmt.Println(test.querystring)
				t.Errorf("Query %v matched unexpected UUID %v", test.querystring, doc.UUID)
				continue
			} else {
				expectedMatches[doc.UUID] = true
			}
		}

		for uuid, covered := range expectedMatches {
			if !covered {
				t.Errorf("Query %v did not match expected UUID %v", test.querystring, uuid)
			}
		}
	}
}
