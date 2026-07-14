package main

import (
	"testing"
	"time"

	"github.com/discruter/scratchpad/internal/assert"
)

func TestHumanDate(t *testing.T) {
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2024, 3, 17, 10, 15, 0, 0, time.UTC),
			want: "17 Mar 2024 at 10:15",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "IST",
			tm:   time.Date(2026, 10, 10, 5, 30, 0, 0, time.FixedZone("IST", -1*(5*60*60+30*60))),
			want: "10 Oct 2026 at 11:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)
			assert.Equal(t, hd, tt.want)
		})
	}
}
