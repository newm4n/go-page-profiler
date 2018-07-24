package go_page_profiler

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"
)

var loadProfilerInstance *LoadProfiler

var ProfileChannel chan *ProfileEntry

func addProfileEntry(c chan *ProfileEntry) {
	for {
		select {
		case pe := <-c:
			ProfilerInstance().addEntry(pe)
		}
	}
}

func CurrentTimeMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

type LoadProfiler struct {
	FastestProfile *Profile            `json:"FastestProfile,omitempty"`
	SlowestProfile *Profile            `json:"SlowestProfile,omitempty"`
	Profiles       map[string]*Profile `json:"Profiles,omitempty"`
}

func (lp *LoadProfiler) String() string {
	var buffer bytes.Buffer
	for _, val := range lp.Profiles {
		buffer.WriteString(fmt.Sprintf("\n\t%s", val.String()))
	}
	return fmt.Sprintf("Fastest : %s\n"+
		"Slowest : %s\n%s",
		lp.FastestProfile.String(),
		lp.SlowestProfile.String(),
		buffer.String())
}

func ProfilerInstance() *LoadProfiler {
	if loadProfilerInstance == nil {
		loadProfilerInstance = &LoadProfiler{Profiles: make(map[string]*Profile)}
		ProfileChannel = make(chan *ProfileEntry)
		go addProfileEntry(ProfileChannel)
		log.Printf("Profiler channel created...")
	}
	return loadProfilerInstance
}

type Profile struct {
	ProfileKey        string
	Count             int64
	TotalTime         int64
	LongestTimeStamp  int64
	LongestTime       int64
	ShortestTimeStamp int64
	ShortestTime      int64
	AverageTime       int64
}

func (p *Profile) String() string {
	return fmt.Sprintf("ProfileKey = %s, Count = %d, TotalTime = %d, "+
		"LongestTimeStamp = %d, LongestTime = %d, ShortestTimeStamp = %d, "+
		"ShortestTime = %d, AverageTime = %d",
		p.ProfileKey, p.Count, p.TotalTime, p.LongestTimeStamp,
		p.LongestTime, p.ShortestTimeStamp, p.ShortestTime, p.AverageTime)
}

type ProfileEntry struct {
	ProfileKey string
	Begin      int64
	End        int64
}

func (pe *ProfileEntry) String() string {
	return fmt.Sprintf("ProfileKey = %s, Begin = %d, End = %d", pe.ProfileKey, pe.Begin, pe.End)
}

func BeginProfile(profileKey string) *ProfileEntry {
	pe := ProfileEntry{
		ProfileKey: profileKey,
		Begin:      CurrentTimeMillis(),
	}
	return &pe
}

func EndProfile(entry *ProfileEntry) {
	entry.End = CurrentTimeMillis()
	ProfileChannel <- entry
}

func (lp *LoadProfiler) addEntry(entry *ProfileEntry) {
	totTime := entry.End - entry.Begin
	if totTime < 0 {
		totTime = -totTime
	}
	var profile *Profile
	if lp.FastestProfile == nil {
		profile = &Profile{
			ProfileKey:        entry.ProfileKey,
			Count:             1,
			TotalTime:         totTime,
			LongestTime:       totTime,
			ShortestTime:      totTime,
			LongestTimeStamp:  CurrentTimeMillis(),
			ShortestTimeStamp: CurrentTimeMillis(),
			AverageTime:       totTime,
		}
		lp.Profiles[entry.ProfileKey] = profile
		lp.FastestProfile = profile
		lp.SlowestProfile = profile
	} else {
		profile = lp.Profiles[entry.ProfileKey]
		if profile != nil {
			profile.Count++
			profile.TotalTime += totTime
			profile.AverageTime = profile.TotalTime / profile.Count
			if profile.ShortestTime > totTime {
				profile.ShortestTime = totTime
				profile.ShortestTimeStamp = CurrentTimeMillis()
			}
			if profile.LongestTimeStamp < totTime {
				profile.LongestTime = totTime
				profile.LongestTimeStamp = CurrentTimeMillis()
			}
		} else {
			profile = &Profile{
				ProfileKey:        entry.ProfileKey,
				Count:             1,
				TotalTime:         totTime,
				LongestTime:       totTime,
				ShortestTime:      totTime,
				LongestTimeStamp:  CurrentTimeMillis(),
				ShortestTimeStamp: CurrentTimeMillis(),
				AverageTime:       totTime,
			}
			lp.Profiles[entry.ProfileKey] = profile
		}
		if profile.AverageTime < lp.FastestProfile.AverageTime {
			lp.FastestProfile = profile
		}
		if profile.AverageTime > lp.SlowestProfile.AverageTime {
			lp.SlowestProfile = profile
		}
	}
}

func LoadProfileFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ProfilerInstance()
		entry := BeginProfile(r.URL.Path)
		next.ServeHTTP(w, r)
		EndProfile(entry)
	})
}
