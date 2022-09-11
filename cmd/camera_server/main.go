package main

import (
	"camera_server/pkg/gst"
	"camera_server/pkg/gst/elements"
	"camera_server/pkg/signal"
	"fmt"
	"github.com/pion/webrtc/v3"
	"time"
)

// TODO the gst library is most likely full of memory leaks, fix

func main() {
	// WebRTC configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.1.google.com:19302"},
			},
		},
	}

	// create RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	panicIfError(err)

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection state has changed to %s \n", connectionState.String())
	})

	// create video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
		MimeType: "video/vp8",
	}, "video", "mainStream")

	_, err = peerConnection.AddTrack(videoTrack)
	panicIfError(err)

	// offer
	offer := webrtc.SessionDescription{}
	signal.Decode(signal.MustReadStdin(), &offer)

	err = peerConnection.SetRemoteDescription(offer)
	panicIfError(err)

	answer, err := peerConnection.CreateAnswer(nil)
	panicIfError(err)

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	err = peerConnection.SetLocalDescription(answer)
	panicIfError(err)

	<-gatherComplete

	fmt.Println(signal.Encode(*peerConnection.LocalDescription()))

	// Start gstreamer
	gst.Init()

	source, err := elements.NewRtspSource("testSource", "rtsp://rtsp.stream/pattern")
	//source, err := elements.NewRtspSource("testSource", "rtsp://admin:L2793C70@10.22.240.53:554/cam/realmonitor?channel=1&subtype=0&proto=Onvif")
	panicIfError(err)

	queue, err := elements.NewQueue("testQueue")
	panicIfError(err)

	depay, err := elements.NewRtpH264Depay("testDepay")
	//depay, err := elements.NewRtpH265Depay("testDepay")
	panicIfError(err)

	parse, err := elements.NewH264Parse("testParse")
	//parse, err := elements.NewH265Parse("testParse")
	panicIfError(err)

	tee, err := elements.NewTee("testTee")

	localQueue, err := elements.NewQueue("localQueue")
	panicIfError(err)

	decode, err := elements.NewAvDecH264("testDec")
	//decode, err := elements.NewAvDecH265("testDec")
	panicIfError(err)

	vp8encode, err := elements.NewVp8Enc("testenc")
	panicIfError(err)

	vp8dec, err := elements.NewVp8Dec("testdecvp8")
	panicIfError(err)

	sink, err := elements.NewAutoVideoSink("testSink")
	panicIfError(err)

	webRtcQueue, err := elements.NewQueue("webRtcQueue")

	webrtcSink, err := elements.NewWebRtcSink("testWebRTC", videoTrack)

	pipeline, err := gst.NewGstPipeline("test-pipeline")
	panicIfError(err)

	pipeline.AddElement(source)
	pipeline.AddElement(queue)
	pipeline.AddElement(depay)
	pipeline.AddElement(parse)
	pipeline.AddElement(tee)
	pipeline.AddElement(localQueue)
	pipeline.AddElement(decode)
	pipeline.AddElement(vp8encode)
	pipeline.AddElement(vp8dec)
	pipeline.AddElement(sink)
	pipeline.AddElement(webRtcQueue)
	pipeline.AddElement(webrtcSink)

	err = gst.LinkElements(queue, depay)
	panicIfError(err)
	err = gst.LinkElements(depay, parse)
	panicIfError(err)
	err = gst.LinkElements(parse, decode)
	panicIfError(err)
	err = gst.LinkElements(decode, vp8encode)
	panicIfError(err)
	err = gst.LinkElements(vp8encode, tee)
	panicIfError(err)
	err = gst.LinkElements(tee, webRtcQueue)
	panicIfError(err)
	err = gst.LinkElements(tee, localQueue)
	panicIfError(err)
	err = gst.LinkElements(localQueue, vp8dec)
	panicIfError(err)
	err = gst.LinkElements(vp8dec, sink)
	panicIfError(err)
	err = gst.LinkElements(webRtcQueue, webrtcSink)
	panicIfError(err)

	padAddedHandler := func(newPad gst.Pad) {
		format, err := (newPad).Format(0)
		panicIfError(err)

		encoding, err := format.QueryStringProperty("encoding-name")
		panicIfError(err)

		if encoding != "H264" {
			return
		}

		sinkPad, err := queue.QueryPadByName("sink")
		panicIfError(err)

		err = gst.LinkPads(newPad, &sinkPad)
		panicIfError(err)

		println("Linked pads!")

	}

	source.OnPadAdded(padAddedHandler)

	time.Sleep(500 * time.Millisecond)

	err = pipeline.SetState(gst.PLAYING)
	panicIfError(err)

	bus, err := pipeline.Bus()
	panicIfError(err)

	for {
		msg, err := bus.PopMessageWithFilter(gst.ERROR | gst.END_OF_STREAM)
		// If there's an error, there's no message to process
		if err == nil {

			switch msg.Type {
			case gst.ERROR:
				println("Error, exiting...")
				debug, err := msg.ParseAsError()
				panicIfError(err)
				println(debug)
				return
			case gst.END_OF_STREAM:
				println("end of stream")
				return

			default:
				panic(fmt.Errorf("unknown message type %i", msg.Type))

			}

		}
		//time.Sleep(25 * time.Millisecond)
	}

}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
