package database

import (
	"os"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/bojand/ghz/web/model"
	"github.com/stretchr/testify/assert"
)

func TestDatabase_Report(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
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

		r.ErrorDist = map[string]int{
			"rpc error: code = Internal desc = Internal error.":            3,
			"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 2}

		r.StatusCodeDist = map[string]int{
			"OK":               195,
			"Internal":         3,
			"DeadlineExceeded": 2}

		r.LatencyDistribution = []*runner.LatencyDistribution{
			{
				Percentage: 25,
				Latency:    time.Duration(1 * time.Millisecond),
			},
			{
				Percentage: 50,
				Latency:    time.Duration(5 * time.Millisecond),
			},
			{
				Percentage: 75,
				Latency:    time.Duration(10 * time.Millisecond),
			},
			{
				Percentage: 90,
				Latency:    time.Duration(15 * time.Millisecond),
			},
			{
				Percentage: 95,
				Latency:    time.Duration(20 * time.Millisecond),
			},
			{
				Percentage: 99,
				Latency:    time.Duration(25 * time.Millisecond),
			},
		}

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
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
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

		r.ErrorDist = map[string]int{
			"rpc error: code = Internal desc = Internal error.":            1,
			"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 4}

		r.StatusCodeDist = map[string]int{
			"OK":               195,
			"Internal":         1,
			"DeadlineExceeded": 4}

		r.LatencyDistribution = []*runner.LatencyDistribution{
			{
				Percentage: 25,
				Latency:    time.Duration(2 * time.Millisecond),
			},
			{
				Percentage: 50,
				Latency:    time.Duration(6 * time.Millisecond),
			},
			{
				Percentage: 75,
				Latency:    time.Duration(11 * time.Millisecond),
			},
			{
				Percentage: 90,
				Latency:    time.Duration(16 * time.Millisecond),
			},
			{
				Percentage: 95,
				Latency:    time.Duration(21 * time.Millisecond),
			},
			{
				Percentage: 99,
				Latency:    time.Duration(27 * time.Millisecond),
			},
		}

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
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
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

		r.ErrorDist = map[string]int{
			"rpc error: code = Internal desc = Internal error.":            2,
			"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 2}

		r.StatusCodeDist = map[string]int{
			"OK":               396,
			"Internal":         2,
			"DeadlineExceeded": 2}

		r.LatencyDistribution = []*runner.LatencyDistribution{
			{
				Percentage: 25,
				Latency:    time.Duration(3 * time.Millisecond),
			},
			{
				Percentage: 50,
				Latency:    time.Duration(7 * time.Millisecond),
			},
			{
				Percentage: 75,
				Latency:    time.Duration(12 * time.Millisecond),
			},
			{
				Percentage: 90,
				Latency:    time.Duration(17 * time.Millisecond),
			},
			{
				Percentage: 95,
				Latency:    time.Duration(22 * time.Millisecond),
			},
			{
				Percentage: 99,
				Latency:    time.Duration(30 * time.Millisecond),
			},
		}

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
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})

	t.Run("FindReportByID", func(t *testing.T) {
		r, err := db.FindReportByID(rid)

		assert.NoError(t, err)
		assert.NotNil(t, r)

		assert.Equal(t, pid, r.ProjectID)
		assert.NotZero(t, r.CreatedAt)
		assert.NotZero(t, r.UpdatedAt)
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

		assert.NotNil(t, r.LatencyDistribution)
		assert.Len(t, r.LatencyDistribution, 6)
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 25,
			Latency:    time.Duration(1 * time.Millisecond),
		}, r.LatencyDistribution[0])
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 99,
			Latency:    time.Duration(25 * time.Millisecond),
		}, r.LatencyDistribution[5])
	})

	t.Run("FindReportByID 2", func(t *testing.T) {
		r, err := db.FindReportByID(rid2)

		assert.NoError(t, err)
		assert.NotNil(t, r)

		assert.Equal(t, pid2, r.ProjectID)
		assert.NotZero(t, r.CreatedAt)
		assert.NotZero(t, r.UpdatedAt)
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

		assert.NotNil(t, r.LatencyDistribution)
		assert.Len(t, r.LatencyDistribution, 6)
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 25,
			Latency:    time.Duration(2 * time.Millisecond),
		}, r.LatencyDistribution[0])
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 99,
			Latency:    time.Duration(27 * time.Millisecond),
		}, r.LatencyDistribution[5])
	})

	t.Run("FindReportByID 3", func(t *testing.T) {
		r, err := db.FindReportByID(rid3)

		assert.NoError(t, err)
		assert.NotNil(t, r)

		assert.Equal(t, pid2, r.ProjectID)
		assert.NotZero(t, r.CreatedAt)
		assert.NotZero(t, r.UpdatedAt)
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

		assert.NotNil(t, r.LatencyDistribution)
		assert.Len(t, r.LatencyDistribution, 6)
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 25,
			Latency:    time.Duration(3 * time.Millisecond),
		}, r.LatencyDistribution[0])
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 99,
			Latency:    time.Duration(30 * time.Millisecond),
		}, r.LatencyDistribution[5])
	})

	t.Run("FindReportByID missing", func(t *testing.T) {
		r, err := db.FindReportByID(123432)

		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("FindPreviousReport", func(t *testing.T) {
		r, err := db.FindPreviousReport(rid3)

		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, rid2, r.ID)
	})

	t.Run("FindPreviousReport invalid id", func(t *testing.T) {
		r, err := db.FindPreviousReport(12345)

		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("FindPreviousReport no previous", func(t *testing.T) {
		r, err := db.FindPreviousReport(rid)

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

	t.Run("ListReportsForProject by pid", func(t *testing.T) {
		list, err := db.ListReportsForProject(pid, 10, 0, "id", "asc")

		assert.NoError(t, err)
		assert.Len(t, list, 1)

		assert.Equal(t, uint(1), list[0].ID)
	})

	t.Run("ListReportsForProject by pid unknown", func(t *testing.T) {
		list, err := db.ListReportsForProject(112311, 10, 0, "id", "asc")

		assert.NoError(t, err)
		assert.Len(t, list, 0)
	})

	t.Run("ListReportsForProject by pid 2 desc", func(t *testing.T) {
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

	t.Run("ListReportsForProject by pid 2 date asc", func(t *testing.T) {
		list, err := db.ListReportsForProject(pid2, 10, 0, "date", "asc")

		assert.NoError(t, err)
		assert.Len(t, list, 2)

		assert.Equal(t, uint(2), list[0].ID)
		assert.Equal(t, uint(3), list[1].ID)
	})

	t.Run("ListReportsForProject default to id", func(t *testing.T) {
		list, err := db.ListReportsForProject(pid, 10, 0, "asdf", "asc")

		assert.NoError(t, err)
		assert.Len(t, list, 1)

		assert.Equal(t, uint(1), list[0].ID)
	})

	t.Run("ListReportsForProject default to asc", func(t *testing.T) {
		list, err := db.ListReportsForProject(pid, 10, 0, "id", "asdf")

		assert.NoError(t, err)
		assert.Len(t, list, 1)

		assert.Equal(t, uint(1), list[0].ID)
	})

	t.Run("FindLatestReportForProject", func(t *testing.T) {
		r, err := db.FindLatestReportForProject(pid2)

		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, rid3, r.ID)
	})

	t.Run("FindLatestReportForProject invalid id", func(t *testing.T) {
		r, err := db.FindLatestReportForProject(12345)

		assert.NoError(t, err)
		assert.Nil(t, r)
	})

	t.Run("DeleteReport()", func(t *testing.T) {
		p := &model.Report{}
		p.ID = rid3

		err := db.DeleteReport(p)

		assert.NoError(t, err)

		r2 := new(model.Report)
		err = db.DB.First(r2, rid3).Error

		assert.Error(t, err)
	})

	t.Run("DeleteReportBulk()", func(t *testing.T) {
		ids := []uint{rid, 123, rid2}

		n, err := db.DeleteReportBulk(ids)

		assert.NoError(t, err)
		assert.Equal(t, 2, n)

		r2 := new(model.Report)
		err = db.DB.First(r2, rid).Error
		assert.Error(t, err)

		err = db.DB.First(r2, rid2).Error
		assert.Error(t, err)
	})
}
