package generator

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
)

func TestID_ShortForm(t *testing.T) {
	tests := []struct {
		name string
		i    ID
		want Short
	}{
		{
			"simple 0 to ascii",
			ID{0, 0, 0, 0, 0, 0, 0, 0},
			Short{48, 48, 48, 48, 48},
		},
		{
			"simple F to ascii",
			ID{255, 255, 255, 255, 255, 255, 255, 255},
			Short{70, 70, 70, 70, 70},
		},
		{
			"last 5 unique",
			ID{0, 0, 0, 0, 0, 127, 255, 255},
			Short{70, 70, 70, 70, 70},
		},
		{
			"last 5 random",
			ID{0, 0, 0, 0, 0, 127, 127, 127},
			Short{70, 55, 70, 55, 70},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.ShortForm(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ID.ShortForm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShort_String(t *testing.T) {
	got := Short{'h', 'e', 'l', 'l', 'o'}.String()
	want := "hello"

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Short.String() = %v, want %v", got, want)
	}
}

func TestID_String(t *testing.T) {
	got := ID{0,0,0,0,255,255,255,255}.String()
	want := "00000000FFFFFFFF"

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ID.String() = %v, want %v", got, want)
	}
}

func helperHexID(h string) ID {
	id, err := NewIDFromHex(h)
	if err != nil {
		panic(err)
	}

	return id
}

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		gen  	*Generator
		want    []ID
		wantErr bool
	}{
		{
			"1 ID",
			NewGenerator(1, RandomReader(rand.New(rand.NewSource(1)).Read)),
			[]ID{
				helperHexID("52FDFC072182654F"),
			},
			false,
		},
		{
			"5 IDs",
			NewGenerator(5, RandomReader(rand.New(rand.NewSource(1)).Read)),
			[]ID{
				helperHexID("52FDFC072182654F"),
				helperHexID("163F5F0F9A621D72"),
				helperHexID("9566C74D10037C4D"),
				helperHexID("7BBB0407D1E2C649"),
				helperHexID("81855AD8681D0D86"),
			},
			false,
		},
		{
			"Error from random",
			NewGenerator(1, RandomReader(func(bytes []byte) (int, error) {
				return 0, errors.New("oops")
			})),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.gen
			got, err := g.Generate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Generate() got = %v, want %v", got, tt.want)
			}
		})
	}
}