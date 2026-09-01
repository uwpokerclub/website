package services

import "testing"

func TestCalculateTiePoints_Singletons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		place int
		want  int
	}{
		{"place_1", 1, 116},
		{"place_2", 2, 99},
		{"place_3", 3, 89},
		{"place_10", 10, 59},
		{"place_25", 25, 36},
		{"place_50", 50, 18},
		{"place_75", 75, 8},
		{"place_100", 100, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CalculateTiePoints(100, tt.place, tt.place, 1.0); got != tt.want {
				t.Errorf("CalculateTiePoints(100, %d, %d, 1.0) = %d, want %d", tt.place, tt.place, got, tt.want)
			}
		})
	}
}

func TestCalculateTiePoints_TieGroups(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		from, to int
		want     int
	}{
		{"top_three", 1, 3, 101},
		{"mid_pack", 50, 60, 16},
		{"long_tail", 41, 100, 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CalculateTiePoints(100, tt.from, tt.to, 1.0); got != tt.want {
				t.Errorf("CalculateTiePoints(100, %d, %d, 1.0) = %d, want %d", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestCalculateTiePoints_ProductionShape(t *testing.T) {
	t.Parallel()

	if got := CalculateTiePoints(328, 1, 1, 1.0); got != 146 {
		t.Errorf("place 1 = %d, want 146", got)
	}
	if got := CalculateTiePoints(328, 86, 86, 1.0); got != 34 {
		t.Errorf("place 86 = %d, want 34", got)
	}
	if got := CalculateTiePoints(328, 87, 328, 1.0); got != 14 {
		t.Errorf("tie 87-328 = %d, want 14", got)
	}
}

func TestCalculateTiePoints_EdgeCases(t *testing.T) {
	t.Parallel()

	if got := CalculateTiePoints(1, 1, 1, 1.0); got != 1 {
		t.Errorf("N=1 place 1 = %d, want 1 (ln(1) = 0)", got)
	}
	if got := CalculateTiePoints(50, 1, 50, 1.0); got != 25 {
		t.Errorf("whole field tied, N=50, group 1-50 = %d, want 25", got)
	}
	if got := CalculateTiePoints(100, 1, 1, 2.0); got != 232 {
		t.Errorf("multiplier applied after averaging = %d, want 232", got)
	}
}
