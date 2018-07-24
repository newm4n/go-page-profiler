package go_page_profiler

import (
	"testing"
	"time"
)

func TestProfilerInstance(t *testing.T) {
	ProfilerInstance()
	pe1 := BeginProfile("test1")
	time.Sleep(2 * time.Second)
	pe2 := BeginProfile("test2")
	time.Sleep(2 * time.Second)
	pe3 := BeginProfile("test3")
	time.Sleep(2 * time.Second)
	EndProfile(pe1)
	EndProfile(pe2)
	pe4 := BeginProfile("test3")
	time.Sleep(500 * time.Millisecond)
	EndProfile(pe3)
	EndProfile(pe4)
	if len(ProfilerInstance().Profiles) != 3 {
		t.Fatal("Profile is not 3")
	}
	println(ProfilerInstance().String())
	if ProfilerInstance().FastestProfile.ProfileKey != "test3" {
		t.Fatal("Fastest is not 3")
	}
	if ProfilerInstance().SlowestProfile.ProfileKey != "test1" {
		t.Fatal("Slowest is not 1")
	}
}