package rsql

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gocraft/dbr"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/require"
)

type CustomInt int

// TestRSQL :
func TestRSQL(t *testing.T) {
	{
		var i struct {
			Name        string    `rsql:"n,filter,sort,allow=eq|gt|gte"`
			Status      string    `rsql:"status,filter,sort"`
			PtrStr      *string   `rsql:"text,filter,sort"`
			No          int       `rsql:"no,filter,sort,column=No2"`
			Int         CustomInt `rsql:"int,filter"`
			SubmittedAt time.Time `rsql:"submittedAt,filter"`
			CreatedAt   time.Time `rsql:"createdAt,sort"`
		}

		p := MustNew(i)
		param, err := p.ParseQuery(`filter=int>10;status=eq="111";no=gt=1991;text==null&sort=status,-no&limit=100&page=2`)
		require.NoError(t, err)
		require.NotNil(t, param)
		require.True(t, len(param.Filters) > 0)
		require.True(t, len(param.Sorts) > 0)
		require.Equal(t, uint(100), param.Limit)

		param, err = p.ParseQuery(`filter=(int>10;status=eq="111";no=gt=1991;text==null)&sort=status,-no&limit=100`)
		require.NoError(t, err)
		require.NotNil(t, param)

		param, err = p.ParseQuery(`filter=(status=="APPROVED";submittedAt>="2019-12-22T16:00:00Z";submittedAt<="2019-12-31T15:59:59Z")&sort=-createdAt&limit=10`)
		require.NoError(t, err)
		require.NotNil(t, param)
		// log.Println("Filters :", param.Filters)

		param, err = p.ParseQuery(`filter=(submittedAt>='2019-12-22T16:00:00Z';submittedAt<='2019-12-31T15:59:59Z';status=="APPROVED")&limit=100`)
		require.NoError(t, err)
		require.NotNil(t, param)
		// log.Println("Filters :", param.Filters)
	}

	{
		var i struct {
			Flag        bool      `rsql:"flag,filter"`
			Status      string    `rsql:"status,filter,sort,allow=eq|gt|gte"`
			SubmittedAt time.Time `rsql:"submittedAt,filter,sort"`
		}
		p := MustNew(i)
		param, err := p.ParseQuery(`filter=status=eq="approved";flag=ne=false;submittedAt>="2019-12-22T16:00:00Z";submittedAt<="2019-12-31T15:59:59Z"&sort=status&limit=10`)
		require.NoError(t, err)
		require.NotNil(t, param)
	}

	{
		var i struct {
			Title            string    `rsql:"title,filter"`
			Audience         string    `rsql:"audience,filter"`
			Status           string    `rsql:"status,filter,sort,allow=eq|gt|gte"`
			ScheduleDateTime time.Time `rsql:"scheduleDateTime,filter,sort"`
		}
		p := MustNew(i)
		param, err := p.ParseQuery(`filter=(audience=="CUSTOMIZED";status=="PENDING";scheduleDateTime>='2020-01-14T16:00:00Z';scheduleDateTime<='2020-01-20T15:59:59Z';title=like="testing%25")`)
		require.NoError(t, err)
		require.NotNil(t, param)
	}

	{
		var i struct {
			Name             string    `rsql:"name,sort"`
			Status           string    `rsql:"status,filter,sort,allow=eq|gt|gte"`
			ScheduleDateTime time.Time `rsql:"scheduleDateTime,filter,sort"`
		}

		p := MustNew(i)
		query := `filter=&sort=name,-status&limit=10&page=2`
		param, err := p.ParseQuery(query)
		require.NoError(t, err)
		require.Equal(t, uint(10), param.Offset)

		/*
			actions.Find().
				From("table").
				Where(
					expr.Equal("key", "value"),
				).
				OrderBy(
					expr.Asc("F1"),
					expr.Asc("F2"),
					expr.Desc("Status"),
				).
				Limit(10).
				Offset(2 * 10)
		*/

	}
}

func TestNoRefType(t *testing.T) {
	p := MustNew(nil)
	param, err := p.ParseQuery(`filter=int>10;status=eq="111";no=gt=1991;text==null&sort=status,-no&limit=100&page=2`)
	require.NoError(t, err)
	require.NotNil(t, param)
	require.True(t, len(param.Filters) > 0)
	require.True(t, len(param.Sorts) > 0)
	require.Equal(t, uint(100), param.Limit)
}

func TestJSONTags(t *testing.T) {
	{
		var i struct {
			Name        string    `json:"n"`
			Status      string    `json:"status"`
			PtrStr      *string   `json:"text"`
			No          int       `json:"no"`
			Int         CustomInt `json:"int"`
			SubmittedAt time.Time `json:"submitted_at"`
			CreatedAt   time.Time `json:"created_at"`
		}
		p := MustNew(&i)
		param, err := p.ParseQuery(`filter=int>10;status=eq="111";no=gt=1991;text==null&sort=status,-no&limit=100&page=2`)
		require.NoError(t, err)
		require.NotNil(t, param)
		require.True(t, len(param.Filters) > 0)
		require.True(t, len(param.Sorts) > 0)
		require.Equal(t, uint(100), param.Limit)
		require.Len(t, param.Filters, 4)
		require.Len(t, param.Sorts, 2)
	}
}

func TestSqlTypes(t *testing.T) {
	type todo struct {
		ID         int64          `json:"id" pk:"true" gorm:"primaryKey"` // id
		Name       dbr.NullString `json:"name"`                           // name
		IsComplete dbr.NullBool   `json:"is_complete"`                    // is_complete
		IsDeleted  dbr.NullBool   `json:"is_deleted"`                     // is_deleted
		CreatedBy  dbr.NullInt64  `json:"created_by"`                     // created_by
		UpdatedBy  dbr.NullInt64  `json:"updated_by"`                     // updated_by
		CreatedAt  dbr.NullTime   `json:"created_at"`                     // created_at
		UpdatedAt  dbr.NullTime   `json:"updated_at"`                     // updated_at
	}

	p := MustNew(todo{})
	param, err := p.ParseQuery(`filter=name=like=foo&sort=updated_at`)
	require.NoError(t, err)
	require.NotNil(t, param)
	require.True(t, len(param.Filters) > 0)
	require.True(t, len(param.Sorts) > 0)
}

func TestParsing(t *testing.T) {
	p := MustNew(nil)

	// list from examples of java parser
	// https://github.com/jirutka/rsql-parser
	queryList := `
	filter=name=="Kill Bill";year=gt=2003
	filter=name=="Kill Bill" and year>2003
	filter=genres=in=(sci-fi,action);(director=='Christopher Nolan',actor==*Bale);year=ge=2000
	filter=genres=in=(sci-fi,action) and (director=='Christopher Nolan' or actor==*Bale) and year>=2000
	filter=director.lastName==Nolan;year=ge=2000;year=lt=2010
	filter=director.lastName==Nolan and year>=2000 and year<2010
	filter=genres=in=(sci-fi,action);genres=out=(romance,animated,horror),director==Que*Tarantino
	filter=genres=in=(sci-fi,action) and genres=out=(romance,animated,horror) or director==Que*Tarantino
	`
	queries := strings.Split(queryList, "\n")
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		fmt.Println(q)
		param, err := p.ParseQuery(q)
		pretty.Println(param)
		require.NoError(t, err)
		require.NotNil(t, param)
		require.NotEmpty(t, param.Filters)
		require.Empty(t, param.Sorts)
	}
}
