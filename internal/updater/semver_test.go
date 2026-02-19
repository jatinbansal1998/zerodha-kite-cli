package updater

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "same", a: "v1.2.3", b: "v1.2.3", want: 0},
		{name: "newer patch", a: "v1.2.4", b: "v1.2.3", want: 1},
		{name: "older minor", a: "v1.1.9", b: "v1.2.0", want: -1},
		{name: "release beats prerelease", a: "v1.2.3", b: "v1.2.3-rc.1", want: 1},
		{name: "prerelease compare", a: "v1.2.3-rc.2", b: "v1.2.3-rc.1", want: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CompareVersions(tc.a, tc.b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestCompareVersionsRejectsInvalid(t *testing.T) {
	if _, err := CompareVersions("dev", "v1.2.3"); err == nil {
		t.Fatalf("expected error for invalid version")
	}
}
