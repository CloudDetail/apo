package clickhouse

import (
	"log"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func Test_chRepo_GetFieldValues(t *testing.T) {
	now := time.Now()
	repo := NewTestRepo(t)

	filters, err := repo.GetAvailableFilterKey(now.Add(-24*time.Hour), time.Now(), false)
	if err != nil {
		t.Error(err.Error())
	}

	for _, filter := range filters {
		options, err := repo.GetFieldValues("", &filter, now.Add(-24*time.Hour), time.Now())
		if err != nil {
			log.Println(err.Error())
		} else {
			log.Printf("options: %+v", options)
		}
	}
}

func Test_chRepo_GetFieldValuesWithSearchText(t *testing.T) {
	now := time.Now()
	repo := NewTestRepo(t)

	options, err := repo.GetFieldValues("fa", &request.SpanTraceFilter{
		Key:         "container_id",
		ParentField: request.PF_Labels,
		DataType:    request.StringColumn,
	}, now.Add(-24*time.Hour), time.Now())
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("options: %+v", options)
	}
}

func Test_chRepo_GetTracePageList(t *testing.T) {
	ch := NewTestRepo(t)

	type args struct {
		req *request.GetTracePageListRequest
	}
	tests := []struct {
		name    string
		args    args
		want    []QueryTraceResult
		want1   int64
		wantErr bool
	}{
		{
			name: "TestOpEqual",
			args: args{
				req: &request.GetTracePageListRequest{
					StartTime:   1725497048000000,
					EndTime:     1725500632000000,
					Service:     "ts-station-service",
					EndPoint:    "",
					Instance:    "",
					NodeName:    "",
					ContainerId: "",
					Pid:         0,
					TraceId:     "",
					PageNum:     1,
					PageSize:    10,
					Filters: []*request.SpanTraceFilter{
						{
							Key:         "node_name",
							ParentField: request.PF_Labels,
							DataType:    request.StringColumn,
							Operation:   request.OpEqual,
							Value: []string{
								"large-server-6",
							},
						},
					},
				},
			},
			want:    []QueryTraceResult{},
			want1:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ch.GetTracePageList(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("chRepo.GetTracePageList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("chRepo.GetTracePageList() got = %v, want %v", got, tt.want)
			// }
			// if got1 != tt.want1 {
			// 	t.Errorf("chRepo.GetTracePageList() got1 = %v, want %v", got1, tt.want1)
			// }
			if len(got) > 0 {
				t.Logf("(TOTAL: %d)got: %+v", got1, got[0])
			}
		})
	}
}

func Test_chRepo_UpdateFilterKey(t *testing.T) {
	repo := NewTestRepo(t)
	now := time.Now()

	filters, err := repo.GetAvailableFilterKey(now.Add(-24*time.Hour), time.Now(), false)
	if err != nil {
		t.Error(err.Error())
	}

	for _, filter := range filters {
		log.Printf("filter: %v", filter)
	}
}
