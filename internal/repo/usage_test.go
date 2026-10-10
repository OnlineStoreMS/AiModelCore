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
	}); err != nil {
		t.Fatal(err)
	}
	out, err := r.TodayUsage(1, time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if out.Calls != 1 || out.Tokens != 12 || out.Errors != 0 {
		t.Fatalf("usage %+v", out)
	}
}
