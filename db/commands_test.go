package db

import (
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/Feresey/banana_bot/logging"
	"github.com/Feresey/banana_bot/model"
)

func TestMain(m *testing.M) {
	ddb, err := sql.Open("postgres", "user=lol dbname=test sslmode=disable")
	if err != nil {
		panic(err)
	}
	db = ddb
	log = logging.NewLogger("test")
	log.Quite = true
	defer db.Close()
	os.Exit(m.Run())
}

func TestWarn(t *testing.T) {
	_, _ = db.Exec(`delete from warn`)
	type args struct {
		person *model.Person
		add    bool
	}
	tests := []struct {
		name      string
		args      args
		wantTotal int
		wantErr   bool
	}{
		{
			name:      "warn new",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: true},
			wantTotal: 1,
		},
		{
			name:      "warn second",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: true},
			wantTotal: 2,
		},
		{
			name:      "unwarn exist",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: false},
			wantTotal: 1,
		},
		{
			name:      "unwarn exist twice",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: false},
			wantTotal: 0,
		},
		{
			name:      "unwarn zero",
			args:      args{person: &model.Person{ChatID: 1, UserID: 1}, add: false},
			wantTotal: 0,
		},
		{
			name:      "unwarn new",
			args:      args{person: &model.Person{ChatID: 1, UserID: 2}, add: false},
			wantTotal: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, err := Warn(tt.args.person, tt.args.add)
			if (err != nil) != tt.wantErr {
				t.Errorf("Warn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("Warn() = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func TestReport(t *testing.T) {
	_, _ = db.Exec(`delete from admins`)
	_, _ = db.Exec(
		`INSERT INTO admins
		VALUES (1,1,true),(1,2,false),(1,3,false),(1,5,true), (2,1,true),(2,2,false)`)
	type args struct {
		chatID int64
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{
			name: "one",
			args: args{chatID: 1},
			want: []int{1, 5},
		},
		{
			name: "two",
			args: args{chatID: 2},
			want: []int{1},
		},
		{
			name: "not exists",
			args: args{chatID: 3},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Report(tt.args.chatID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Report() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Report() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetChatList(t *testing.T) {
	_, _ = db.Exec(`delete from admins`)
	_, _ = db.Exec(
		`INSERT INTO admins
		VALUES (1,1,true),(1,2,false),(1,3,false),(1,5,true), (2,1,true),(2,2,false)`)
	tests := []struct {
		name    string
		want    []int64
		wantErr bool
	}{
		{
			name: "",
			want: []int64{1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChatList()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChatList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChatList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAdmins(t *testing.T) {
	_, _ = db.Exec(`delete from admins`)
	_, _ = db.Exec(
		`INSERT INTO admins
		VALUES (1,1,true),(1,2,false),(1,3,false),(1,5,true), (2,1,true),(2,2,false)`)
	type args struct {
		chatid int64
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{
			name: "one",
			args: args{chatid: 1},
			want: []int{1, 2, 3, 5},
		},
		{
			name: "two",
			args: args{chatid: 2},
			want: []int{1, 2},
		},
		{
			name: "nil",
			args: args{chatid: 3},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAdmins(tt.args.chatid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAdmins() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAdmins() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetChatsForAdmin(t *testing.T) {
	_, _ = db.Exec(`delete from admins`)
	_, _ = db.Exec(
		`INSERT INTO admins
		VALUES	(1,1,true),(1,2,false),(1,3,false),(1,5,true), 
				(2,1,true),(2,2,false)`)
	type args struct {
		userid int
	}
	tests := []struct {
		name    string
		args    args
		want    []int64
		wantErr bool
	}{
		{
			name: "one in two",
			args: args{userid: 1},
			want: []int64{1, 2},
		},
		{
			name: "one in one",
			args: args{userid: 3},
			want: []int64{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChatsForAdmin(tt.args.userid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChatsForAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChatsForAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAdmins(t *testing.T) {
	type args struct {
		chatid int64
		pipls  []int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "init",
			args: args{chatid: 1, pipls: []int{1, 2, 3}},
		},
		{
			name: "add new",
			args: args{chatid: 1, pipls: []int{1, 2, 3, 4, 5, 6}},
		},
		{
			name: "remove some",
			args: args{chatid: 1, pipls: []int{1, 3, 5, 6}},
		},
		{
			name: "all new",
			args: args{chatid: 1, pipls: []int{8, 9, 0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetAdmins(tt.args.chatid, tt.args.pipls); (err != nil) != tt.wantErr {
				t.Errorf("SetAdmins() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
