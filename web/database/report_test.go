package database

import (
	"os"
	"testing"
	"time"

	"github.com/bojand/ghz/web/model"
	"github.com/stretchr/testify/assert"
)

func TestDatabase_Report(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	var rid, rid2, rid3, pid, pid2 uint

	t.Run("test new", func(t *testing.T) {
		p := model.Project{
			Name:        "Test Proj 111 ",
			Description: "Test Description Asdf ",
		}

		r := model.Report{
			Project:   &p,
			Name:      "Test report",
			EndReason: "normal",
			Date:      time.Date(2018, 12, 1, 8, 0, 0, 0, time.UTC),
			Count:     200,
			Total:     time.Duration(2 * time.Second),
			Average:   time.Duration(10 * time.Millisecond),
			Fastest:   time.Duration(1 * time.Millisecond),
			Slowest:   time.Duration(100 * time.Millisecond),
			Rps:       2000,
		}

		r.Options = &model.Options{
			Name:        "Test report",
			Call:        "helloworld.Greeter.SayHello",
			Proto:       "../../testdata/greeter.proto",
			Host:        "0.0.0.0:50051",
			N:           200,
			C:           50,
			Timeout:     time.Duration(20 * time.Second),
			DialTimeout: time.Duration(10 * time.Second),
			CPUs:        8,
			Insecure:    true,
			Data:        map[string]string{"name": "Joe"},
			Metadata:    &map[string]string{"token": "abc123", "request-id": "12345"},
		}

		r.ErrorDist = map[string]int{
			"rpc error: code = Internal desc = Internal error.":            3,
			"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 2}

		r.StatusCodeDist = map[string]int{
			"OK":               195,
			"Internal":         3,
			"DeadlineExceeded": 2}

		// r.LatencyDistribution = []*runner.LatencyDistribution{
		// 	&runner.LatencyDistribution{
		// 		Percentage: 25,
		// 		Latency:    time.Duration(1 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 50,
		// 		Latency:    time.Duration(5 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 75,
		// 		Latency:    time.Duration(10 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 90,
		// 		Latency:    time.Duration(15 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 95,
		// 		Latency:    time.Duration(20 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 99,
		// 		Latency:    time.Duration(25 * time.Millisecond),
		// 	},
		// }

		// r.Histogram = []*runner.Bucket{
		// 	&runner.Bucket{
		// 		Mark:      0.01,
		// 		Count:     1,
		// 		Frequency: 0.005,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.02,
		// 		Count:     10,
		// 		Frequency: 0.01,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.03,
		// 		Count:     50,
		// 		Frequency: 0.1,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.05,
		// 		Count:     60,
		// 		Frequency: 0.15,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.1,
		// 		Count:     15,
		// 		Frequency: 0.07,
		// 	},
		// }

		err := db.CreateReport(&r)

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)

		pid = p.ID
		rid = r.ID

		p2 := new(model.Project)
		err = db.DB.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, p.Name, p2.Name)
		assert.Equal(t, "Test Description Asdf", p2.Description)
		assert.Equal(t, model.StatusOK, p2.Status)
		assert.NotNil(t, p2.CreatedAt)
		assert.NotNil(t, p2.UpdatedAt)
		assert.Nil(t, p2.DeletedAt)
	})

	t.Run("test new 2", func(t *testing.T) {
		p := model.Project{
			Name:        "Test Proj 222 ",
			Description: "Test Description project 2 ",
		}

		r := model.Report{
			Project:   &p,
			Name:      "Test report 2",
			EndReason: "canceled",
			Date:      time.Date(2018, 12, 2, 9, 0, 0, 0, time.UTC),
			Count:     300,
			Total:     time.Duration(3 * time.Second),
			Average:   time.Duration(11 * time.Millisecond),
			Fastest:   time.Duration(2 * time.Millisecond),
			Slowest:   time.Duration(120 * time.Millisecond),
			Rps:       2222,
		}

		r.Options = &model.Options{
			Name:        "Test report 2",
			Call:        "helloworld.Greeter.SayHello",
			Proto:       "../../testdata/greeter.proto",
			Host:        "0.0.0.0:50051",
			N:           300,
			C:           50,
			Timeout:     time.Duration(20 * time.Second),
			DialTimeout: time.Duration(10 * time.Second),
			CPUs:        8,
			Insecure:    true,
			Data:        map[string]string{"name": "Kate"},
			Metadata:    &map[string]string{"token": "foo123", "request-id": "321"},
		}

		r.ErrorDist = map[string]int{
			"rpc error: code = Internal desc = Internal error.":            1,
			"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 4}

		r.StatusCodeDist = map[string]int{
			"OK":               195,
			"Internal":         1,
			"DeadlineExceeded": 4}

		// r.LatencyDistribution = []*runner.LatencyDistribution{
		// 	&runner.LatencyDistribution{
		// 		Percentage: 25,
		// 		Latency:    time.Duration(2 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 50,
		// 		Latency:    time.Duration(6 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 75,
		// 		Latency:    time.Duration(11 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 90,
		// 		Latency:    time.Duration(16 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 95,
		// 		Latency:    time.Duration(21 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 99,
		// 		Latency:    time.Duration(27 * time.Millisecond),
		// 	},
		// }

		// r.Histogram = []*runner.Bucket{
		// 	&runner.Bucket{
		// 		Mark:      0.02,
		// 		Count:     2,
		// 		Frequency: 0.006,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.02,
		// 		Count:     10,
		// 		Frequency: 0.01,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.03,
		// 		Count:     50,
		// 		Frequency: 0.1,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.05,
		// 		Count:     60,
		// 		Frequency: 0.15,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.15,
		// 		Count:     19,
		// 		Frequency: 0.08,
		// 	},
		// }

		err := db.CreateReport(&r)

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)

		pid2 = p.ID
		rid2 = r.ID

		p2 := new(model.Project)
		err = db.DB.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, p.Name, p2.Name)
		assert.Equal(t, "Test Description project 2", p2.Description)
		assert.Equal(t, model.StatusOK, p2.Status)
		assert.NotNil(t, p2.CreatedAt)
		assert.NotNil(t, p2.UpdatedAt)
		assert.Nil(t, p2.DeletedAt)
	})

	t.Run("test new for project 2", func(t *testing.T) {
		r := model.Report{
			ProjectID: pid2,
			Name:      "Test report 3",
			EndReason: "normal",
			Date:      time.Date(2018, 12, 3, 10, 0, 0, 0, time.UTC),
			Count:     400,
			Total:     time.Duration(3 * time.Second),
			Average:   time.Duration(11 * time.Millisecond),
			Fastest:   time.Duration(2 * time.Millisecond),
			Slowest:   time.Duration(120 * time.Millisecond),
			Rps:       2567,
		}

		r.Options = &model.Options{
			Name:        "Test report 3",
			Call:        "helloworld.Greeter.SayHello",
			Proto:       "../../testdata/greeter.proto",
			Host:        "0.0.0.0:50051",
			N:           400,
			C:           40,
			Timeout:     time.Duration(20 * time.Second),
			DialTimeout: time.Duration(10 * time.Second),
			CPUs:        8,
			Binary:      true,
			Insecure:    false,
			Data:        map[string]string{"name": "Bob"},
			Metadata:    &map[string]string{"token": "bar321", "request-id": "555"},
		}

		r.ErrorDist = map[string]int{
			"rpc error: code = Internal desc = Internal error.":            2,
			"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 2}

		r.StatusCodeDist = map[string]int{
			"OK":               396,
			"Internal":         2,
			"DeadlineExceeded": 2}

		// r.LatencyDistribution = []*runner.LatencyDistribution{
		// 	&runner.LatencyDistribution{
		// 		Percentage: 25,
		// 		Latency:    time.Duration(3 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 50,
		// 		Latency:    time.Duration(7 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 75,
		// 		Latency:    time.Duration(12 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 90,
		// 		Latency:    time.Duration(17 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 95,
		// 		Latency:    time.Duration(22 * time.Millisecond),
		// 	},
		// 	&runner.LatencyDistribution{
		// 		Percentage: 99,
		// 		Latency:    time.Duration(30 * time.Millisecond),
		// 	},
		// }

		// r.Histogram = []*runner.Bucket{
		// 	&runner.Bucket{
		// 		Mark:      0.03,
		// 		Count:     3,
		// 		Frequency: 0.007,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.02,
		// 		Count:     10,
		// 		Frequency: 0.01,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.05,
		// 		Count:     60,
		// 		Frequency: 0.15,
		// 	},
		// 	&runner.Bucket{
		// 		Mark:      0.17,
		// 		Count:     22,
		// 		Frequency: 0.11,
		// 	},
		// }

		err := db.CreateReport(&r)

		assert.NoError(t, err)
		assert.NotZero(t, r.ID)

		rid3 = r.ID

		p2 := new(model.Project)
		err = db.DB.First(p2, pid2).Error

		assert.NoError(t, err)
		assert.Equal(t, "Test Proj 222", p2.Name)
		assert.Equal(t, "Test Description project 2", p2.Description)
		assert.Equal(t, model.StatusOK, p2.Status)
		assert.NotNil(t, p2.CreatedAt)
		assert.NotNil(t, p2.UpdatedAt)
		assert.Nil(t, p2.DeletedAt)
	})

	t.Run("FindReportByID", func(t *testing.T) {
		r := new(model.Report)
		r, err = db.FindReportByID(rid)

		assert.NoError(t, err)
		assert.NotNil(t, r)

		assert.Equal(t, pid, r.ProjectID)
		assert.NotNil(t, r.CreatedAt)
		assert.NotNil(t, r.UpdatedAt)
		assert.Nil(t, r.DeletedAt)
		assert.Equal(t, model.StatusOK, r.Status)

		assert.Equal(t, "Test report", r.Name)
		assert.Equal(t, "normal", r.EndReason)
		assert.NotZero(t, r.Date)
		assert.Equal(t, 2000.0, r.Rps)

		assert.Equal(t, 3, r.ErrorDist["rpc error: code = Internal desc = Internal error."])
		assert.Equal(t, 2, r.ErrorDist["rpc error: code = DeadlineExceeded desc = Deadline exceeded."])

		assert.Equal(t, 195, r.StatusCodeDist["OK"])
		assert.Equal(t, 3, r.StatusCodeDist["Internal"])
		assert.Equal(t, 2, r.StatusCodeDist["DeadlineExceeded"])

		assert.Equal(t, "Test report", r.Options.Name)
		assert.Equal(t, "helloworld.Greeter.SayHello", r.Options.Call)
		assert.Equal(t, "../../testdata/greeter.proto", r.Options.Proto)
		assert.Equal(t, "0.0.0.0:50051", r.Options.Host)
		assert.Equal(t, uint(200), r.Options.N)
		assert.Equal(t, uint(50), r.Options.C)
		assert.Equal(t, time.Duration(20*time.Second), r.Options.Timeout)
		assert.Equal(t, time.Duration(10*time.Second), r.Options.DialTimeout)
		assert.Equal(t, map[string]interface{}{"name": "Joe"}, r.Options.Data)
		assert.Equal(t, &map[string]string{"token": "abc123", "request-id": "12345"}, r.Options.Metadata)
		assert.Equal(t, false, r.Options.Binary)
		assert.Equal(t, true, r.Options.Insecure)
		assert.Equal(t, 8, r.Options.CPUs)

		// assert.NotNil(t, r.LatencyDistribution)
		// assert.Len(t, r.LatencyDistribution, 6)
		// assert.Equal(t, &runner.LatencyDistribution{
		// 	Percentage: 25,
		// 	Latency:    time.Duration(1 * time.Millisecond),
		// }, r.LatencyDistribution[0])
		// assert.Equal(t, &runner.LatencyDistribution{
		// 	Percentage: 99,
		// 	Latency:    time.Duration(25 * time.Millisecond),
		// }, r.LatencyDistribution[5])

		// assert.NotNil(t, r.Histogram)
		// assert.Len(t, r.Histogram, 5)
		// assert.Equal(t, &runner.Bucket{
		// 	Mark:      0.01,
		// 	Count:     1,
		// 	Frequency: 0.005,
		// }, r.Histogram[0])
		// assert.Equal(t, &runner.Bucket{
		// 	Mark:      0.1,
		// 	Count:     15,
		// 	Frequency: 0.07,
		// }, r.Histogram[4])
	})

	t.Run("FindReportByID 2", func(t *testing.T) {
		r := new(model.Report)
		r, err = db.FindReportByID(rid2)

		assert.NoError(t, err)
		assert.NotNil(t, r)

		assert.Equal(t, pid2, r.ProjectID)
		assert.NotNil(t, r.CreatedAt)
		assert.NotNil(t, r.UpdatedAt)
		assert.Nil(t, r.DeletedAt)
		assert.Equal(t, model.StatusOK, r.Status)
		assert.Equal(t, 2222.0, r.Rps)

		assert.Equal(t, "Test report 2", r.Name)
		assert.Equal(t, "canceled", r.EndReason)
		assert.NotZero(t, r.Date)

		assert.Equal(t, 1, r.ErrorDist["rpc error: code = Internal desc = Internal error."])
		assert.Equal(t, 4, r.ErrorDist["rpc error: code = DeadlineExceeded desc = Deadline exceeded."])

		assert.Equal(t, 195, r.StatusCodeDist["OK"])
		assert.Equal(t, 1, r.StatusCodeDist["Internal"])
		assert.Equal(t, 4, r.StatusCodeDist["DeadlineExceeded"])

		assert.Equal(t, "Test report 2", r.Options.Name)
		assert.Equal(t, "helloworld.Greeter.SayHello", r.Options.Call)
		assert.Equal(t, "../../testdata/greeter.proto", r.Options.Proto)
		assert.Equal(t, "0.0.0.0:50051", r.Options.Host)
		assert.Equal(t, uint(300), r.Options.N)
		assert.Equal(t, uint(50), r.Options.C)
		assert.Equal(t, time.Duration(20*time.Second), r.Options.Timeout)
		assert.Equal(t, time.Duration(10*time.Second), r.Options.DialTimeout)
		assert.Equal(t, map[string]interface{}{"name": "Kate"}, r.Options.Data)
		assert.Equal(t, &map[string]string{"token": "foo123", "request-id": "321"}, r.Options.Metadata)
		assert.Equal(t, false, r.Options.Binary)
		assert.Equal(t, true, r.Options.Insecure)
		assert.Equal(t, 8, r.Options.CPUs)

		// assert.NotNil(t, r.LatencyDistribution)
		// assert.Len(t, r.LatencyDistribution, 6)
		// assert.Equal(t, &runner.LatencyDistribution{
		// 	Percentage: 25,
		// 	Latency:    time.Duration(2 * time.Millisecond),
		// }, r.LatencyDistribution[0])
		// assert.Equal(t, &runner.LatencyDistribution{
		// 	Percentage: 99,
		// 	Latency:    time.Duration(27 * time.Millisecond),
		// }, r.LatencyDistribution[5])

		// assert.NotNil(t, r.Histogram)
		// assert.Len(t, r.Histogram, 5)
		// assert.Equal(t, &runner.Bucket{
		// 	Mark:      0.02,
		// 	Count:     2,
		// 	Frequency: 0.006,
		// }, r.Histogram[0])
		// assert.Equal(t, &runner.Bucket{
		// 	Mark:      0.15,
		// 	Count:     19,
		// 	Frequency: 0.08,
		// }, r.Histogram[4])
	})

	t.Run("FindReportByID 3", func(t *testing.T) {
		r := new(model.Report)
		r, err = db.FindReportByID(rid3)

		assert.NoError(t, err)
		assert.NotNil(t, r)

		assert.Equal(t, pid2, r.ProjectID)
		assert.NotNil(t, r.CreatedAt)
		assert.NotNil(t, r.UpdatedAt)
		assert.Nil(t, r.DeletedAt)
		assert.Equal(t, model.StatusOK, r.Status)
		assert.Equal(t, 2567.0, r.Rps)

		assert.Equal(t, "Test report 3", r.Name)
		assert.Equal(t, "normal", r.EndReason)
		assert.NotZero(t, r.Date)

		assert.Equal(t, 2, r.ErrorDist["rpc error: code = Internal desc = Internal error."])
		assert.Equal(t, 2, r.ErrorDist["rpc error: code = DeadlineExceeded desc = Deadline exceeded."])

		assert.Equal(t, 396, r.StatusCodeDist["OK"])
		assert.Equal(t, 2, r.StatusCodeDist["Internal"])
		assert.Equal(t, 2, r.StatusCodeDist["DeadlineExceeded"])

		assert.Equal(t, "Test report 3", r.Options.Name)
		assert.Equal(t, "helloworld.Greeter.SayHello", r.Options.Call)
		assert.Equal(t, "../../testdata/greeter.proto", r.Options.Proto)
		assert.Equal(t, "0.0.0.0:50051", r.Options.Host)
		assert.Equal(t, uint(400), r.Options.N)
		assert.Equal(t, uint(40), r.Options.C)
		assert.Equal(t, time.Duration(20*time.Second), r.Options.Timeout)
		assert.Equal(t, time.Duration(10*time.Second), r.Options.DialTimeout)
		assert.Equal(t, map[string]interface{}{"name": "Bob"}, r.Options.Data)
		assert.Equal(t, &map[string]string{"token": "bar321", "request-id": "555"}, r.Options.Metadata)
		assert.Equal(t, true, r.Options.Binary)
		assert.Equal(t, false, r.Options.Insecure)
		assert.Equal(t, 8, r.Options.CPUs)

		// assert.NotNil(t, r.LatencyDistribution)
		// assert.Len(t, r.LatencyDistribution, 6)
		// assert.Equal(t, &runner.LatencyDistribution{
		// 	Percentage: 25,
		// 	Latency:    time.Duration(3 * time.Millisecond),
		// }, r.LatencyDistribution[0])
		// assert.Equal(t, &runner.LatencyDistribution{
		// 	Percentage: 99,
		// 	Latency:    time.Duration(30 * time.Millisecond),
		// }, r.LatencyDistribution[5])

		// assert.NotNil(t, r.Histogram)
		// assert.Len(t, r.Histogram, 4)
		// assert.Equal(t, &runner.Bucket{
		// 	Mark:      0.03,
		// 	Count:     3,
		// 	Frequency: 0.007,
		// }, r.Histogram[0])
		// assert.Equal(t, &runner.Bucket{
		// 	Mark:      0.17,
		// 	Count:     22,
		// 	Frequency: 0.11,
		// }, r.Histogram[3])
	})

	t.Run("FindReportByID missing", func(t *testing.T) {
		r := new(model.Report)
		r, err = db.FindReportByID(123432)

		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("CountReports", func(t *testing.T) {
		count, err := db.CountReports()

		assert.NoError(t, err)
		assert.Equal(t, uint(3), count)
	})

	t.Run("CountReportsForProject", func(t *testing.T) {
		count, err := db.CountReportsForProject(pid)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), count)
	})

	t.Run("CountReportsForProject 2", func(t *testing.T) {
		count, err := db.CountReportsForProject(pid2)

		assert.NoError(t, err)
		assert.Equal(t, uint(2), count)
	})

	t.Run("CountReportsForProject unknown", func(t *testing.T) {
		count, err := db.CountReportsForProject(43232)

		assert.NoError(t, err)
		assert.Equal(t, uint(0), count)
	})

	t.Run("ListReports", func(t *testing.T) {
		list, err := db.ListReports(10, 0, "id", "asc")

		assert.NoError(t, err)
		assert.Len(t, list, 3)

		assert.Equal(t, uint(1), list[0].ID)
		assert.Equal(t, uint(3), list[2].ID)
	})

	t.Run("ListReports date desc", func(t *testing.T) {
		list, err := db.ListReports(10, 0, "date", "desc")

		assert.NoError(t, err)
		assert.Len(t, list, 3)

		assert.Equal(t, uint(3), list[0].ID)
		assert.Equal(t, uint(1), list[2].ID)
	})

	t.Run("ListReports by pid", func(t *testing.T) {
		list, err := db.ListReportsForProject(pid, 10, 0, "id", "asc")

		assert.NoError(t, err)
		assert.Len(t, list, 1)

		assert.Equal(t, uint(1), list[0].ID)
	})

	t.Run("ListReports by pid unknown", func(t *testing.T) {
		list, err := db.ListReportsForProject(112311, 10, 0, "id", "asc")

		assert.NoError(t, err)
		assert.Len(t, list, 0)
	})

	t.Run("ListReports by pid 2 desc", func(t *testing.T) {
		list, err := db.ListReportsForProject(pid2, 10, 0, "id", "desc")

		assert.NoError(t, err)
		assert.Len(t, list, 2)

		assert.Equal(t, uint(3), list[0].ID)
		assert.Equal(t, uint(2), list[1].ID)
	})

	t.Run("ListReports date desc", func(t *testing.T) {
		list, err := db.ListReports(10, 0, "date", "desc")

		assert.NoError(t, err)
		assert.Len(t, list, 3)

		assert.Equal(t, uint(3), list[0].ID)
		assert.Equal(t, uint(1), list[2].ID)
	})

	t.Run("ListReports by pid 2 date asc", func(t *testing.T) {
		list, err := db.ListReportsForProject(pid2, 10, 0, "date", "asc")

		assert.NoError(t, err)
		assert.Len(t, list, 2)

		assert.Equal(t, uint(2), list[0].ID)
		assert.Equal(t, uint(3), list[1].ID)
	})
}
