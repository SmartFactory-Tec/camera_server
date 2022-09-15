package main

import "C"
import (
	"camera_server/pkg/gst"
	"camera_server/pkg/gst/elements"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type WebRtcSink struct {
	elements.AppSink
}

func NewWebRtcSink(name string, track *webrtc.TrackLocalStaticSample) (WebRtcSink, error) {
	createdAppSink, err := elements.NewAppSink(name)
	if err != nil {
		return WebRtcSink{}, err
	}

	createdAppSink.SetProperty("emit-signals", true)

	createdAppSink.OnNewSample(func(newSample gst.Sample) {
		buffer := newSample.Buffer()
		data := buffer.Bytes()
		duration := buffer.Duration()
		if err := track.WriteSample(media.Sample{
			Data:     data,
			Duration: duration,
		}); err != nil {
			panic(err)
		}
	})

	return WebRtcSink{createdAppSink}, nil
}

//
////export newSampleHandler
//func newSampleHandler(element *C.BaseElement, *C.void) {
//	C.g_signal_emit_by_name()
//}
