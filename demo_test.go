package demo_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

type Country struct {
	Code string `db:"code"`
	Name string `db:"name"`
}

type City struct {
	Code        string `db:"code"`
	Name        string `db:"name"`
	CountryCode string `db:"countryCode"`
}

func Test(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.Nil(t, err)
	assert.NotNil(t, db)

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbmap.TraceOn("[gorp]", log.New(os.Stdout, "", log.Lmicroseconds))
	dbmap.AddTable(Country{}).SetKeys(false, "Code")
	dbmap.AddTable(City{}).SetKeys(false, "Code")
	assert.Nil(t, dbmap.CreateTablesIfNotExists())

	tx, err := dbmap.Begin()
	assert.Nil(t, err)
	assert.Nil(t, tx.Insert(&Country{Code: "PT", Name: "Portugal"}))
	assert.Nil(t, tx.Insert(&Country{Code: "SP", Name: "Spain"}))
	assert.Nil(t, tx.Insert(&City{Code: "OPO", Name: "Porto", CountryCode: "PT"}))
	assert.Nil(t, tx.Insert(&City{Code: "LIS", Name: "Lisbon", CountryCode: "PT"}))
	assert.Nil(t, tx.Insert(&City{Code: "BAR", Name: "Barcelona", CountryCode: "SP"}))
	assert.Nil(t, tx.Insert(&City{Code: "MAD", Name: "Madrid", CountryCode: "SP"}))
	assert.Nil(t, tx.Commit())

	countries := []Country{}
	dbmap.Select(&countries, "select * from Country order by Code")
	assert.Equal(t, 2, len(countries))
	assert.Equal(t, "PT", countries[0].Code)
	assert.Equal(t, "Portugal", countries[0].Name)
	assert.Equal(t, "SP", countries[1].Code)
	assert.Equal(t, "Spain", countries[1].Name)

	countries = []Country{}
	countryPks := []string{"PT", "SP"}
	//dbmap.Select(&countries, "select * from Country where Code in (?, ?) order by Code", "PT", "SP")
	dbmap.Select(&countries, fmt.Sprintf("select * from Country where Code in (%s) order by Code", questionMarks(len(countryPks))), "PT", "SP")
	assert.Equal(t, 2, len(countries))
	assert.Equal(t, "PT", countries[0].Code)
	assert.Equal(t, "Portugal", countries[0].Name)
	assert.Equal(t, "SP", countries[1].Code)
	assert.Equal(t, "Spain", countries[1].Name)

	tx, err = dbmap.Begin()
	assert.Nil(t, err)
	count, err := tx.Update(&City{Code: "LIS", Name: "Lisboa"})
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Nil(t, tx.Commit())

	lisbon, err := dbmap.Get(City{}, "LIS")
	assert.Nil(t, err)
	assert.Equal(t, "LIS", lisbon.(*City).Code)
	assert.Equal(t, "Lisboa", lisbon.(*City).Name)
	assert.Equal(t, "", lisbon.(*City).CountryCode) // lost CountryCode after last Update

	tx, err = dbmap.Begin()
	assert.Nil(t, err)
	count, err = tx.Delete(&City{Code: "LIS"})
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Nil(t, tx.Commit())

	lisbon, err = dbmap.Get(City{}, "LIS")
	assert.Nil(t, err)
	assert.Nil(t, lisbon)

	count, err = dbmap.SelectInt("select count(*) from City where Code = ?", "LIS")
	assert.Nil(t, err)
	assert.Equal(t, 0, count)

	count, err = dbmap.SelectInt("select count(*) from City where Code = ?", "OPO")
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func questionMarks(n int) string {
	q := []string{}
	for i := 0; i < n; i++ {
		q = append(q, "?")
	}
	return strings.Join(q, ",")
}
