package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"

type RtspSource struct {
	location string
	*BaseElement
}

func NewRtspSource(name string, location string) (RtspSource, error) {
	createdElement, err := NewGstElement("rtspsrc", name)

	if err != nil {
		return RtspSource{}, err
	}

	createdElement.SetProperty("location", location)

	return RtspSource{location: location, BaseElement: &createdElement}, nil
}
