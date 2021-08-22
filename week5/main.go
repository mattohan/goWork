package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	ReportTypeSuccess = 1
	ReportTypeFailure = 2
	ReportTypeTimeout = 3
	ReportTypeReject = 4
)


type SlideWindow struct {
	WindowStart time.Time
	BucketStartIdx int
	BucketUnit time.Duration
	BucketCount int
	BucketSpan time.Duration
	Buckets []Bucket

	mu sync.Locker
}


type Bucket struct {
	Success int
	Failure int
	Timeout int
	Rejection int
}

func (b *Bucket) Reset() {
	b.Success = 0
	b.Failure = 0
	b.Timeout = 0
	b.Rejection = 0
}

func (b *Bucket) Update(t int) {
	switch t {
	case ReportTypeSuccess:
		b.Success += 1
	case ReportTypeFailure:
		b.Failure += 1
	case ReportTypeTimeout:
		b.Timeout += 1
	case ReportTypeReject:
		b.Rejection += 1
	}
}


func NewSlideWindow(windowStart time.Time, bucketUnit time.Duration, bucketCount int) *SlideWindow {
	buckets := make([]Bucket, bucketCount)

	return &SlideWindow{
		WindowStart: windowStart,
		BucketStartIdx: 0,
		BucketUnit: bucketUnit,
		BucketSpan: time.Duration(bucketCount) * bucketUnit,
		BucketCount: bucketCount,
		Buckets: buckets,
		mu: &sync.Mutex{},
	}
}

func (sw *SlideWindow) Report(reportType int, reportTime time.Time) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	delta := reportTime.Sub(sw.WindowStart)
	if delta < 0 {
		return errors.New("滞后的上报消息，丢弃")
	}

	sw.slide(&delta)

	bucketIdx := sw.BucketStartIdx + int(delta / sw.BucketUnit)
	if bucketIdx >= sw.BucketCount {
		bucketIdx -= sw.BucketCount
	}

	sw.Buckets[bucketIdx].Update(reportType)

	return nil
}

func (sw *SlideWindow) slide(step *time.Duration) {
	for *step >= sw.BucketSpan {
		// 处理过期的统计信息
		sw.process()

		// 更新滑动窗口信息
		sw.WindowStart = sw.WindowStart.Add(sw.BucketUnit)
		sw.BucketStartIdx += 1
		if sw.BucketStartIdx >= sw.BucketCount {
			sw.BucketStartIdx -= sw.BucketCount
		}
		*step = *step - sw.BucketUnit
	}
}


func (sw *SlideWindow) process() {
	bucketHead := &sw.Buckets[sw.BucketStartIdx]
	fmt.Printf("[窗口范围%s-%s]统计计数：\n成功：%d\n失败：%d\n超时：%d\n拒绝：%d\n",
		sw.WindowStart,
		sw.WindowStart.Add(sw.BucketUnit),
		bucketHead.Success,
		bucketHead.Failure,
		bucketHead.Timeout,
		bucketHead.Rejection,
	)
	bucketHead.Reset()
}


func main() {
	sw := NewSlideWindow(time.Now(), time.Second, 10)
	ch := make(chan int)
	go func(){
		for {
			time.Sleep(1 * time.Second)
			sw.Report(ReportTypeSuccess, time.Now())
		}
	}()

	go func(){
		for {
			time.Sleep(2 * time.Second)
			sw.Report(ReportTypeSuccess, time.Now())
		}
	}()

	<- ch
}


