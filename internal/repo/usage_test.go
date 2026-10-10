package repo

import (
	"testing"
	"time"

	"aimodelcore/internal/database"
	"aimodelcore/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestTodayUsage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:aimodel-usage?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	r := New(db)
	if err := r.AddUsage(&model.UsageRecord{
		TenantID: 1, ModelName: "deepseek-chat", CallerApp: "productcore",
		Capability: "chat", TotalTokens: 12, Status: "ok",
		RequestBytes: 100, ResponseBytes: 40, LatencyMS: 800,
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.AddUsage(&model.UsageRecord{
		TenantID: 1, ModelName: "qwen-image-3.0", CallerApp: "productcore",
		Capability: "image", Status: "error",
		RequestBytes: 300, ResponseBytes: 20, LatencyMS: 1200,
	}); err != nil {
		t.Fatal(err)
	}
	out, err := r.TodayUsage(1, time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if out.Calls != 2 || out.Tokens != 12 || out.Errors != 1 {
		t.Fatalf("usage %+v", out)
	}
	if out.RequestBytes != 400 || out.ResponseBytes != 60 || out.LatencySumMS != 2000 || out.LatencyMaxMS != 1200 {
		t.Fatalf("traffic %+v", out)
	}
}
