package util

import (
	"net"
	"strings"
	"testing"
)

func TestGetCurrentFuncName(t *testing.T) {
	tests := []struct {
		actual string
		want   string
	}{
		{GetCurrentFuncName(), "util.TestGetCurrentFuncName"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if tt.actual != tt.want {
				t.Errorf("actual = %v, want %v", tt.actual, tt.want)
			}
		})
	}
}

func TestGetTypeNameByObject(t *testing.T) {
	tests := []struct {
		obj  interface{}
		want string
	}{
		{strings.Builder{}, "strings.Builder"},
		{&strings.Builder{}, "strings.Builder"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := GetTypeNameByObject(tt.obj); got != tt.want {
				t.Errorf("GetTypeNameByObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFuncName(t *testing.T) {
	tests := []struct {
		f    interface{}
		want string
	}{
		{(&strings.Builder{}).String, "strings.Builder.String"},
		{net.Addr.Network, "net.Addr.Network"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := GetFuncName(tt.f); got != tt.want {
				t.Errorf("GetFuncName() = %v, want %v", got, tt.want)
			}
		})
	}
}
