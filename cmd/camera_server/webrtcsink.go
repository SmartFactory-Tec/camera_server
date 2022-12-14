package main

import "C"
import (
	"camera_server/pkg/gst"
	"context"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type WebRtcSink struct {
	gst.AppSink
	track *webrtc.TrackLocalStaticSample
}

func NewWebRtcSink(name string, track *webrtc.TrackLocalStaticSample) (WebRtcSink, error) {
	createdAppSink, err := gst.NewAppSink(name)
	if err != nil {
		return WebRtcSink{}, err
	}

	return WebRtcSink{createdAppSink, track}, nil
}

func (w *WebRtcSink) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// If context is done stop loop
			return
		default:
			sample, err := w.PullSample()

			if err != nil {
				// Don't process if nothing is available yet
				continue
			}

			buffer := sample.Buffer()
			data := buffer.Bytes()
			duration := buffer.Duration()

			if err := w.track.WriteSample(media.Sample{
				Data:     data,
				Duration: duration,
			}); err != nil {
				break
			}
		}

	}

}

//
////export newSampleHandler
//func newSampleHandler(element *C.BaseElement, *C.void) {
//	C.g_signal_emit_by_name()
//}
