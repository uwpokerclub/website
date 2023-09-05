package services

import "testing"

func TestCalculatePoints(t *testing.T) {
	type args struct {
		eventSize        int
		placement        int
		pointsMultiplier float32
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "64_1st",
			args: args{
				eventSize:        64,
				placement:        1,
				pointsMultiplier: 1.0,
			},
			want: 41,
		},
		{
			name: "64_50th",
			args: args{
				eventSize:        64,
				placement:        50,
				pointsMultiplier: 1.0,
			},
			want: 2,
		},
		{
			name: "64_35th",
			args: args{
				eventSize:        64,
				placement:        35,
				pointsMultiplier: 2.0,
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculatePoints(tt.args.eventSize, tt.args.placement, tt.args.pointsMultiplier); got != tt.want {
				t.Errorf("CalculatePoints() = %v, want %v", got, tt.want)
			}
		})
	}
}
