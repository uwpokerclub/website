package services

import "testing"

func TestCalculatePoints(t *testing.T) {
	type args struct {
		eventSize int
		placement int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "64_1st",
			args: args{
				eventSize: 64,
				placement: 1,
			},
			want: 41,
		},
		{
			name: "64_50th",
			args: args{
				eventSize: 64,
				placement: 50,
			},
			want: 2,
		},
		{
			name: "64_35th",
			args: args{
				eventSize: 64,
				placement: 35,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculatePoints(tt.args.eventSize, tt.args.placement); got != tt.want {
				t.Errorf("CalculatePoints() = %v, want %v", got, tt.want)
			}
		})
	}
}
